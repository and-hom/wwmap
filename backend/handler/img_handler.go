package handler

import (
	"github.com/and-hom/wwmap/lib/blob"
	"net/http"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/gorilla/mux"
	"strconv"
	"time"
	"io"
	"image"
	"golang.org/x/image/draw"
	"image/png"
	_ "image/jpeg"
	_ "image/gif"
	"bytes"
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/Sirupsen/logrus"
	"encoding/json"
	"io/ioutil"
	"github.com/and-hom/wwmap/lib/util"
)

const (
	PREVIEW_MAX_HEIGHT = 200
	PREVIEW_MAX_WIDTH = 300

	BIG_IMG_MAX_HEIGHT = 1000
	BIG_IMG_MAX_WIDTH = 1500
)

type ImgHandler struct {
	App
	ImgStorage        blob.BlobStorage
	PreviewImgStorage blob.BlobStorage
};

func (this *ImgHandler) Init() {
	this.Register("/spot/{spotId}/img", HandlerFunctions{Get: this.GetImages, Post:this.Upload, Put:this.Upload})
	this.Register("/spot/{spotId}/img/{imgId}", HandlerFunctions{Get:this.GetImage, Delete: this.Delete})
	this.Register("/spot/{spotId}/img/{imgId}/preview", HandlerFunctions{Get: this.GetImagePreview})
	this.Register("/spot/{spotId}/img/{imgId}/enabled", HandlerFunctions{Post:this.SetEnabled})
	this.Register("/spot/{spotId}/preview", HandlerFunctions{Get: this.GetPreview, Post:this.SetPreview, Delete:this.DropPreview})
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
		OnError500(w, err, "Can not get image")
		return
	}
	defer r.Close()
	io.Copy(w, r)
}

func (this *ImgHandler) GetImagePreview(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	r, err := this.PreviewImgStorage.Read(storageKeyById(pathParams["imgId"]))
	if err != nil {
		OnError500(w, err, "Can not get image")
		return
	}
	defer r.Close()
	io.Copy(w, r)
}

func (this *ImgHandler) Upload(w http.ResponseWriter, req *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN, dao.EDITOR) {
		return
	}

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

	img, err := this.ImgDao.InsertLocal(spotId, getImgType(req), dao.IMG_SOURCE_WWMAP, this.ImgUrlBase, this.ImgUrlPreviewBase, time.Now())
	if err != nil {
		OnError500(w, err, "Can not insert")
		return
	}

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

	sourceImage, _, err := image.Decode(f)
	if err != nil {
		OnError500(w, err, "Can not get decode image file")
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

func (this *ImgHandler) Delete(w http.ResponseWriter, req *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN, dao.EDITOR) {
		return
	}

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

	err = this.ImgDao.Remove(imgId, nil)
	if err != nil {
		OnError500(w, err, "Can not delete image from db")
		return
	}

	imgRemoveErr := this.ImgStorage.Remove(imgIdStr)
	previewRemoveErr := this.PreviewImgStorage.Remove(imgIdStr)
	if imgRemoveErr != nil {
		logrus.Errorf("Can not delete image data: ", imgRemoveErr)
	}
	if previewRemoveErr != nil {
		logrus.Errorf("Can not delete image preview: ", previewRemoveErr)
	}

	this.listImagesForSpot(w, spotId, getImgType(req))
}

func (this *ImgHandler) SetEnabled(w http.ResponseWriter, req *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN, dao.EDITOR) {
		return
	}

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

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		OnError500(w, err, "Can not read body")
		return
	}
	enabled := false
	json.Unmarshal(bodyBytes, &enabled)

	err = this.ImgDao.SetEnabled(imgId, enabled)
	if err != nil {
		OnError500(w, err, "Can not set image enables/disabled")
		return
	}

	this.listImagesForSpot(w, spotId, getImgType(req))
}

func (this *ImgHandler) SetPreview(w http.ResponseWriter, req *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN, dao.EDITOR) {
		return
	}

	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		OnError500(w, err, "Can not read body")
		return
	}
	imgId := int64(0)
	json.Unmarshal(bodyBytes, &imgId)

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
		OnError500(w, err, fmt.Sprintf("Can not set preview for spot %d", spotId))
		return
	}

	this.listImagesForSpot(w, spotId, getImgType(req))
}

func (this *ImgHandler) DropPreview(w http.ResponseWriter, req *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN, dao.EDITOR) {
		return
	}

	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = this.ImgDao.DropMainForSpot(spotId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not set preview for spot %d", spotId))
		return
	}
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
		OnError500(w, err, fmt.Sprintf("Can not set preview for spot %d", spotId))
		return
	}
	if !found {
		OnError(w, nil, "No main image set for spot", http.StatusNotFound)
		return
	}

	this.processForWeb(&img)
	this.JsonAnswer(w, img)
}

func (this *ImgHandler) listImagesForSpot(w http.ResponseWriter, spotId int64, _type dao.ImageType) {
	imgs, err := this.ImgDao.List(spotId, 16384, _type, false)
	if err != nil {
		OnError500(w, err, "Can not list images")
		return
	}
	for i := 0; i < len(imgs); i++ {
		this.processForWeb(&imgs[i])
	}

	this.JsonAnswer(w, imgs)
}

func getImgType(req *http.Request) dao.ImageType {
	return dao.GetImgType(req.FormValue("type"))
}