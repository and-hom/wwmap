package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/vodinfo-eye/graduation"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/gorilla/mux"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"golang.org/x/image/draw"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"time"
	"github.com/disintegration/imageorient"
)

const (
	PREVIEW_MAX_HEIGHT = 200
	PREVIEW_MAX_WIDTH  = 300

	BIG_IMG_MAX_HEIGHT = 1000
	BIG_IMG_MAX_WIDTH  = 1500

	MAX_EXIF_SIZE = 32000
)

type ImgHandler struct {
	App
	LevelDao          dao.LevelDao
	LevelSensorDao    dao.LevelSensorDao
	ImgStorage        blob.BlobStorage
	PreviewImgStorage blob.BlobStorage
}

func (this *ImgHandler) Init() {
	this.Register("/{entityType}/{entityId}/img", HandlerFunctions{
		Get: this.GetImages,
		Post: this.ForRoles(this.Upload, dao.ADMIN, dao.EDITOR),
		Put:  this.ForRoles(this.Upload, dao.ADMIN, dao.EDITOR),
	})
	this.Register("/spot/{spotId}/img_ext", HandlerFunctions{
		Post: this.ForRoles(this.AddExternalImage, dao.ADMIN, dao.EDITOR),
		Put:  this.ForRoles(this.AddExternalImage, dao.ADMIN, dao.EDITOR),
	})
	this.Register("/spot/{spotId}/img/{imgId}", HandlerFunctions{
		Get: this.GetImage,
		Delete: this.ForRoles(this.DeleteReturningAll, dao.ADMIN, dao.EDITOR),
	})

	this.Register("/{entityType}/{entityId}/img/{imgId}/preview", HandlerFunctions{
		Get: this.GetImagePreview,
	})
	this.Register("/{entityType}/{entityId}/preview", HandlerFunctions{
		Get: this.GetPreview,
		Post:   this.ForRoles(this.SetPreview, dao.ADMIN, dao.EDITOR),
		Delete: this.ForRoles(this.DropPreview, dao.ADMIN, dao.EDITOR),
	})

	this.Register("/{entityType}/{entityId}/img/{imgId}/enabled", HandlerFunctions{
		Post: this.ForRoles(this.SetEnabledReturningAll, dao.ADMIN, dao.EDITOR),
	})

	this.Register("/spot/{spotId}/img/{imgId}/date", HandlerFunctions{
		Post: this.ForRoles(this.SetDate, dao.ADMIN, dao.EDITOR)})
	this.Register("/spot/{spotId}/img/{imgId}/manual-level", HandlerFunctions{
		Post:   this.ForRoles(this.SetManualLevel, dao.ADMIN, dao.EDITOR),
		Delete: this.ForRoles(this.ResetManualLevel, dao.ADMIN, dao.EDITOR),
	})

	this.Register("/img/{imgId}", HandlerFunctions{
		Delete: this.ForRoles(this.Delete, dao.ADMIN, dao.EDITOR),
	})
	this.Register("/img/{imgId}/enabled", HandlerFunctions{
		Post: this.ForRoles(this.SetEnabled, dao.ADMIN, dao.EDITOR),
	})

	exif.RegisterParsers(mknote.All...)
}

func (this *ImgHandler) GetImages(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.listImagesForSpot(w, spotId, getImgType(req))
}

func (this *ImgHandler) GetImage(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	r, err := this.ImgStorage.Read(storageKeyById(pathParams["imgId"]))
	if err != nil {
		OnError(w, err, "Can not get image", http.StatusNotFound)
		return
	}
	defer r.Close()
	io.Copy(w, r)
}

func (this *ImgHandler) GetImagePreview(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	r, err := this.PreviewImgStorage.Read(storageKeyById(pathParams["imgId"]))
	if err != nil {
		OnError(w, err, "Can not get image", http.StatusNotFound)
		return
	}
	defer r.Close()
	io.Copy(w, r)
}

func (this *ImgHandler) AddExternalImage(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	if spotId <= 0 {
		OnError(w, err, "Can not add image for non existing spot", http.StatusBadRequest)
		return
	}

	data := ExternalImageAddData{}
	if err = json.NewDecoder(req.Body).Decode(&data); err != nil {
		OnError(w, err, "Can't parse request body", http.StatusBadRequest)
		return
	}
	img, err := this.ImgDao.Upsert(dao.Img{
		WwId:            spotId,
		RemoteId:        data.Id,
		Type:            dao.ImageType(data.Type),
		Source:          data.Source,
		MainImage:       false,
		Url:             "",
		PreviewUrl:      "",
		RawUrl:          "",
		Enabled:         true,
		LabelsForSearch: []string{},
		Props:           data.Props,
	})
	if err != nil {
		OnError500(w, err, "Can not insert")
		return
	}
	JsonAnswer(w, img)
}

func (this *ImgHandler) Upload(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	if spotId <= 0 {
		OnError(w, err, "Can not upload image for non existing spot", http.StatusBadRequest)
		return
	}
	imgType := getImgType(req)

	err = req.ParseMultipartForm(128 * 1024 * 1024)
	if err != nil {
		OnError500(w, err, "Can not parse multipart form")
		return
	}
	f, _, err := req.FormFile("file")
	if err != nil {
		OnError500(w, err, "Can not get uploaded file")
		return
	}
	defer f.Close()

	var reader io.Reader = f
	headCacher := util.CreateHeadCachingReader(f, MAX_EXIF_SIZE)
	if imgType == dao.IMAGE_TYPE_IMAGE {
		reader = headCacher
	}

	sourceImage, _, err := imageorient.Decode(reader)
	if err != nil {
		OnError500(w, err, "Can not get decode image file")
		return
	}

	var realDate *time.Time = nil
	if imgType == dao.IMAGE_TYPE_IMAGE {
		realDate = util.GetImageRealDate(bytes.NewReader(headCacher.GetCache().Bytes()))
	}
	level := this.getLevelsForDate(spotId, realDate)
	img, err := this.ImgDao.InsertLocal(imgType, dao.IMG_SOURCE_WWMAP, time.Now(), realDate, level)
	if err != nil {
		OnError500(w, err, "Can not insert")
		return
	}

	previewReader, err := compress(sourceImage, f, PREVIEW_MAX_WIDTH, PREVIEW_MAX_HEIGHT, true)
	if err != nil {
		OnError500(w, err, "Can not compress preview")
		return
	}
	err = this.PreviewImgStorage.Store(storageKey(img), previewReader)
	if err != nil {
		OnError500(w, err, "Can not store preview")
		return
	}

	bigImgReader, err := compress(sourceImage, f, BIG_IMG_MAX_WIDTH, BIG_IMG_MAX_HEIGHT, false)
	if err != nil {
		OnError500(w, err, "Can not compress image")
		return
	}
	err = this.ImgStorage.Store(storageKey(img), bigImgReader)
	if err != nil {
		OnError500(w, err, "Can not store image")
		return
	}
	this.LogUserEvent(req, IMAGE_LOG_ENTRY_TYPE, img.Id, dao.ENTRY_TYPE_CREATE, fmt.Sprintf("%s/%s", img.Source, img.RemoteId))
}

func (this *ImgHandler) getLevelsForDate(spotId int64, date *time.Time) map[string]int8 {
	if date == nil {
		return make(map[string]int8)
	}
	river, err := this.RiverDao.FindForSpot(spotId)
	if err != nil {
		logrus.Errorf("Can not select river for spot: id=%d", spotId)
		return make(map[string]int8)
	} else {
		sensorIds := river.GetSensorIds()
		return graduation.GetLevelBySensors(this.LevelSensorDao, this.LevelDao, sensorIds, *date, 1, -1)
	}
}

func storageKey(img dao.Img) string {
	return fmt.Sprintf("%d.png", img.Id)
}

func storageKeyById(imgId string) string {
	return fmt.Sprintf("%s.png", imgId)
}

func compress(sourceImage image.Image, src io.ReadSeeker, maxW, maxH int, resizeSmallerImages bool) (io.Reader, error) {
	rect, small := util.PreviewRect(sourceImage.Bounds(), maxW, maxH)
	if small && !resizeSmallerImages {
		src.Seek(0, 0)
		return src, nil
	}

	resized := image.NewRGBA(rect)
	draw.ApproxBiLinear.Scale(resized, rect, sourceImage, sourceImage.Bounds(), draw.Over, nil)
	var b bytes.Buffer
	err := png.Encode(&b, resized)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

//Deprecated
func (this *ImgHandler) DeleteReturningAll(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	imgIdStr := pathParams["imgId"]
	imgId, err := strconv.ParseInt(imgIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse img id", http.StatusBadRequest)
		return
	}

	existing, found, err := this.ImgDao.Find(imgId)
	if err != nil {
		OnError500(w, err, "Can't select existing image from database")
		return
	}
	if !found {
		OnError(w, err, "Image does not exist", http.StatusNotFound)
		return
	}

	err = this.ImgDao.Remove(imgId, nil)
	if err != nil {
		OnError500(w, err, "Can not delete image from db")
		return
	}

	if existing.Source == dao.IMG_SOURCE_WWMAP {
		storageKey := storageKeyById(imgIdStr)
		imgRemoveErr := this.ImgStorage.Remove(storageKey)
		previewRemoveErr := this.PreviewImgStorage.Remove(storageKey)
		if imgRemoveErr != nil {
			logrus.Error("Can not delete image data: ", imgRemoveErr)
		}
		if previewRemoveErr != nil {
			logrus.Error("Can not delete image preview: ", previewRemoveErr)
		}
	}

	this.listImagesForSpot(w, spotId, getImgType(req))
	this.LogUserEvent(req, IMAGE_LOG_ENTRY_TYPE, imgId, dao.ENTRY_TYPE_DELETE, "")
}


func (this *ImgHandler) Delete(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	imgIdStr := pathParams["imgId"]
	imgId, err := strconv.ParseInt(imgIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse img id", http.StatusBadRequest)
		return
	}

	existing, found, err := this.ImgDao.Find(imgId)
	if err != nil {
		OnError500(w, err, "Can't select existing image from database")
		return
	}
	if !found {
		OnError(w, err, "Image does not exist", http.StatusNotFound)
		return
	}

	err = this.ImgDao.Remove(imgId, nil)
	if err != nil {
		OnError500(w, err, "Can not delete image from db")
		return
	}

	if existing.Source == dao.IMG_SOURCE_WWMAP {
		storageKey := storageKeyById(imgIdStr)
		imgRemoveErr := this.ImgStorage.Remove(storageKey)
		previewRemoveErr := this.PreviewImgStorage.Remove(storageKey)
		if imgRemoveErr != nil {
			logrus.Error("Can not delete image data: ", imgRemoveErr)
		}
		if previewRemoveErr != nil {
			logrus.Error("Can not delete image preview: ", previewRemoveErr)
		}
	}

	JsonAnswer(w, true)

	this.LogUserEvent(req, IMAGE_LOG_ENTRY_TYPE, imgId, dao.ENTRY_TYPE_DELETE, "")
}

//Deprecated
func (this *ImgHandler) SetEnabledReturningAll(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	imgIdStr := pathParams["imgId"]
	imgId, err := strconv.ParseInt(imgIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse img id", http.StatusBadRequest)
		return
	}

	enabled := false
	body, err := DecodeJsonBody(req, &enabled)
	if err != nil {
		OnError(w, err, "Can not unmarshal request body: "+body, http.StatusBadRequest)
		return
	}

	err = this.ImgDao.SetEnabled(imgId, enabled)
	if err != nil {
		OnError500(w, err, "Can not set image enables/disabled")
		return
	}

	this.listImagesForSpot(w, spotId, getImgType(req))

	this.LogUserEvent(req, IMAGE_LOG_ENTRY_TYPE, imgId, dao.ENTRY_TYPE_MODIFY, fmt.Sprintf("enabled=%t", enabled))
}

func (this *ImgHandler) SetEnabled(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	imgIdStr := pathParams["imgId"]
	imgId, err := strconv.ParseInt(imgIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse img id", http.StatusBadRequest)
		return
	}

	enabled := false
	body, err := DecodeJsonBody(req, &enabled)
	if err != nil {
		OnError(w, err, "Can not unmarshal request body: "+body, http.StatusBadRequest)
		return
	}

	err = this.ImgDao.SetEnabled(imgId, enabled)
	if err != nil {
		OnError500(w, err, "Can not set image enables/disabled")
		return
	}

	JsonAnswer(w, true)

	this.LogUserEvent(req, IMAGE_LOG_ENTRY_TYPE, imgId, dao.ENTRY_TYPE_MODIFY, fmt.Sprintf("enabled=%t", enabled))
}

func (this *ImgHandler) SetDate(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	imgIdStr := pathParams["imgId"]
	imgId, err := strconv.ParseInt(imgIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse img id", http.StatusBadRequest)
		return
	}

	date := util.ZeroDateUTC()
	body, err := DecodeJsonBody(req, &date)
	if err != nil {
		OnError(w, err, "Can not unmarshal request body: "+body, http.StatusBadRequest)
		return
	}

	river, err := this.RiverDao.FindForImage(imgId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not select river for image: id=%d", imgId))
		return
	}

	img, found, err := this.ImgDao.Find(imgId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not select image: id=%d", imgId))
		return
	}
	if !found {
		OnError(w, err, fmt.Sprintf("Can not find image: id=%d", imgId), http.StatusNotFound)
		return
	}

	manualLevel, found := img.Level[dao.IMG_WATER_LEVEL_MANUAL]
	if !found {
		manualLevel = graduation.NO_LEVEL_FOR_DATE
	}

	sensorIds := river.GetSensorIds()
	level := graduation.GetLevelBySensors(this.LevelSensorDao, this.LevelDao, sensorIds, date, 1, manualLevel)

	err = this.ImgDao.SetDateAndLevel(imgId, date, level)
	if err != nil {
		OnError500(w, err, "Can not set image date")
		return
	}

	JsonAnswer(w, level)

	this.LogUserEvent(req, IMAGE_LOG_ENTRY_TYPE, imgId, dao.ENTRY_TYPE_MODIFY, fmt.Sprintf("date=%v", date))
}

func (this *ImgHandler) SetManualLevel(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	imgIdStr := pathParams["imgId"]
	imgId, err := strconv.ParseInt(imgIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse img id", http.StatusBadRequest)
		return
	}

	level := int8(0)
	body, err := DecodeJsonBody(req, &level)
	if err != nil {
		OnError(w, err, "Can not unmarshal request body: "+body, http.StatusBadRequest)
		return
	}

	levels, err := this.ImgDao.SetManualLevel(imgId, level)
	if err != nil {
		OnError500(w, err, "Can not set image manual level")
		return
	}

	JsonAnswer(w, levels)

	this.LogUserEvent(req, IMAGE_LOG_ENTRY_TYPE, imgId, dao.ENTRY_TYPE_MODIFY, fmt.Sprintf("manual level=%v", level))
}

func (this *ImgHandler) ResetManualLevel(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	imgIdStr := pathParams["imgId"]
	imgId, err := strconv.ParseInt(imgIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse img id", http.StatusBadRequest)
		return
	}

	levels, err := this.ImgDao.ResetManualLevel(imgId)
	if err != nil {
		OnError500(w, err, "Can not set image manual level")
		return
	}

	JsonAnswer(w, levels)

	this.LogUserEvent(req, IMAGE_LOG_ENTRY_TYPE, imgId, dao.ENTRY_TYPE_MODIFY, "reset manual level")
}

func (this *ImgHandler) SetPreview(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	imgId := int64(0)
	body, err := DecodeJsonBody(req, &imgId)
	if err != nil {
		OnError(w, err, "Can not unmarshal request body: "+body, http.StatusBadRequest)
	}

	img, found, err := this.ImgDao.Find(imgId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not find image id=%d", imgId))
		return
	}
	if !found {
		OnError(w, nil, fmt.Sprintf("Image id=%d does not exist", imgId), http.StatusBadRequest)
		return
	}

	err = this.ImgDao.SetMain(spotId, img.Id)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not set main image for spot %d", spotId))
		return
	}

	this.listImagesForSpot(w, spotId, getImgType(req))
	this.LogUserEvent(req, IMAGE_LOG_ENTRY_TYPE, imgId, dao.ENTRY_TYPE_MODIFY, "main=true")
}

func (this *ImgHandler) DropPreview(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = this.ImgDao.DropMainForSpot(spotId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Drop main image for spot %d", spotId))
		return
	}
	this.LogUserEvent(req, SPOT_LOG_ENTRY_TYPE, spotId, dao.ENTRY_TYPE_MODIFY, "drop main img")
}

func (this *ImgHandler) GetPreview(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	img, found, err := this.ImgDao.GetMainForSpot(spotId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get main image for spot %d", spotId))
		return
	}
	if !found {
		OnErrorWithCustomLogging(w, nil, "No main image set for spot", http.StatusNotFound, func(msg string) {
			logrus.Debug(msg)
		})
		return
	}

	this.processForWeb(&img)
	JsonAnswer(w, img)
}

func (this *ImgHandler) listImagesForSpot(w http.ResponseWriter, spotId int64, _type dao.ImageType) {
	imgs, err := this.ImgDao.ListExt(spotId, 16384, _type, false)
	if err != nil {
		OnError500(w, err, "Can not list images")
		return
	}
	for i := 0; i < len(imgs); i++ {
		this.processForWeb(&imgs[i].Img)
	}

	JsonAnswer(w, imgs)
}

func getImgType(req *http.Request) dao.ImageType {
	return dao.GetImgType(req.FormValue("type"))
}

type ExternalImageAddData struct {
	Id     string                 `json:"id"`
	Type   string                 `json:"type"`
	Source string                 `json:"source"`
	Props  map[string]interface{} `json:"props"`
}
