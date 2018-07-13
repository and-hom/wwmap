package main

import (
	"github.com/and-hom/wwmap/lib/img_storage"
	"net/http"
	. "github.com/and-hom/wwmap/lib/http"
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
)

const SOURCE string = "wwmap"
const (
	PREVIEW_MAX_HEIGHT = 150
	PREVIEW_MAX_WIDTH = 150
)

type ImgHandler struct {
	Handler
	imgStorage        img_storage.ImgStorage
	previewImgStorage img_storage.ImgStorage
	imgUrlBase        string
	imgUrlPreviewBase string
};

func (this *ImgHandler) GetImages(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, GET, POST, PUT, DELETE)

	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	imgs, err := this.imgDao.List(spotId, 16384)
	if err != nil {
		OnError500(w, err, "Can not list images")
		return
	}
	for i:=0;i<len(imgs);i++ {
		if imgs[i].Source == SOURCE {
			imgs[i].Url = fmt.Sprintf(this.imgUrlBase, imgs[i].Id)
			imgs[i].PreviewUrl = fmt.Sprintf(this.imgUrlPreviewBase, imgs[i].Id)
		}
	}

	this.JsonAnswer(w, imgs)
}

func (this *ImgHandler) GetImage(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, GET, POST, PUT, DELETE)
	pathParams := mux.Vars(req)
	r, err := this.imgStorage.Read(pathParams["imgId"])
	if err != nil {
		OnError500(w, err, "Can not get image")
		return
	}
	defer r.Close()
	io.Copy(w, r)
}

func (this *ImgHandler) GetImagePreview(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, GET, POST, PUT, DELETE)
	pathParams := mux.Vars(req)
	r, err := this.previewImgStorage.Read(pathParams["imgId"])
	if err != nil {
		OnError500(w, err, "Can not get image")
		return
	}
	defer r.Close()
	io.Copy(w, r)
}

func (this *ImgHandler) Upload(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, GET, POST, PUT, DELETE)

	pathParams := mux.Vars(req)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	img, err := this.imgDao.InsertLocal(spotId, SOURCE, this.imgUrlBase, this.imgUrlPreviewBase, time.Now())
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
	this.previewImgStorage.Store(img.IdStr(), &b)


	f.Seek(0, 0)
	err = this.imgStorage.Store(img.IdStr(), f)
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
	CorsHeaders(w, GET, POST, PUT, DELETE)

	pathParams := mux.Vars(req)
	imgId, err := strconv.ParseInt(pathParams["imgId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse img id", http.StatusBadRequest)
		return
	}

	this.imgDao.Remove(imgId)
}