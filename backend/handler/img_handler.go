package handler

import (
	"github.com/and-hom/wwmap/lib/img_storage"
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
	"math"
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/Sirupsen/logrus"
	"encoding/json"
	"io/ioutil"
)

const (
	PREVIEW_MAX_HEIGHT = 150
	PREVIEW_MAX_WIDTH = 150
)

type ImgHandler struct {
	App
	ImgStorage        img_storage.ImgStorage
	PreviewImgStorage img_storage.ImgStorage
	ImgUrlBase        string
	ImgUrlPreviewBase string
};

func (this *ImgHandler) Init(r *mux.Router) {
	this.Register(r, "/spot/{spotId}/img", HandlerFunctions{Get: this.GetImages, Post:this.Upload, Put:this.Upload})
	this.Register(r, "/spot/{spotId}/img/{imgId}", HandlerFunctions{Get:this.GetImage, Delete: this.Delete})
	this.Register(r, "/spot/{spotId}/img/{imgId}/preview", HandlerFunctions{Get: this.GetImagePreview})
	this.Register(r, "/spot/{spotId}/img/{imgId}/enabled", HandlerFunctions{Post:this.SetEnabled})
	this.Register(r, "/spot/{spotId}/preview", HandlerFunctions{Get: this.GetPreview, Post:this.SetPreview, Delete:this.DropPreview})
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
	r, err := this.ImgStorage.Read(pathParams["imgId"])
	if err != nil {
		OnError500(w, err, "Can not get image")
		return
	}
	defer r.Close()
	io.Copy(w, r)
}

func (this *ImgHandler) GetImagePreview(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	r, err := this.PreviewImgStorage.Read(pathParams["imgId"])
	if err != nil {
		OnError500(w, err, "Can not get image")
		return
	}
	defer r.Close()
	io.Copy(w, r)
}

func (this *ImgHandler) Upload(w http.ResponseWriter, req *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN) {
		return
	}

	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	if spotId<=0 {
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
	previewRect := previewRect(sourceImage.Bounds())
	preview := image.NewRGBA(previewRect)
	draw.ApproxBiLinear.Scale(preview, previewRect, sourceImage, sourceImage.Bounds(), draw.Over, nil)
	var b bytes.Buffer
	err = png.Encode(&b, preview)
	if err != nil {
		OnError500(w, err, "Can not store preview")
		return
	}
	this.PreviewImgStorage.Store(img.IdStr(), &b)

	f.Seek(0, 0)
	err = this.ImgStorage.Store(img.IdStr(), f)
	if err != nil {
		OnError500(w, err, "Can not store image")
		return
	}
}

func previewRect(r image.Rectangle) image.Rectangle {
	d := math.Abs(float64(r.Max.X - r.Min.X) / float64(r.Max.Y - r.Min.Y))
	w := PREVIEW_MAX_WIDTH
	h := PREVIEW_MAX_HEIGHT

	if d > 1 {
		h = int(float64(h) / d)
	}
	if d < 1 {
		w = int(float64(w) * d)
	}
	return image.Rect(0, 0, w, h)
}

func (this *ImgHandler) Delete(w http.ResponseWriter, req *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN) {
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

	err = this.ImgDao.Remove(imgId)
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
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN) {
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
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN) {
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

	img,found, err := this.ImgDao.Find(imgId)
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
	if !this.CheckRoleAllowedAndMakeResponse(w, req, dao.ADMIN) {
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

func (this *ImgHandler) processForWeb(img *dao.Img) {
	if img.Source == dao.IMG_SOURCE_WWMAP {
		img.Url = fmt.Sprintf(this.ImgUrlBase, img.Id)
		img.PreviewUrl = fmt.Sprintf(this.ImgUrlPreviewBase, img.Id)
	}
}

func getImgType(req *http.Request) dao.ImageType {
	return dao.GetImgType(req.FormValue("type"))
}