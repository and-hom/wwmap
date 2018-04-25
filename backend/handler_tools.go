package main

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"fmt"
	"io"
	"os"
	"errors"
	. "github.com/and-hom/wwmap/lib/geoparser"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/geo"
	"strconv"
)

func (this *Handler) JsonpAnswer(callback string, object interface{}, _default string) []byte {
	return []byte(callback + "(" + this.JsonStr(object, _default) + ");")
}

func (this *Handler) JsonStr(f interface{}, _default string) string {
	bytes, err := json.Marshal(f)
	if err != nil {
		log.Errorf("Can not serialize object %v: %s", f, err.Error())
		return _default
	}
	return string(bytes)
}

func (this *Handler) CorsGetOptionsStub(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")
	// for cors only
}

func (this *Handler) bboxFormValue(w http.ResponseWriter, req *http.Request) (geo.Bbox, error) {
	bboxStr := req.FormValue("bbox")
	bbox, err := geo.NewBbox(bboxStr)
	if err != nil {
		this.onError(w, err, fmt.Sprintf("Can not parse bbox: %v", bbox), http.StatusBadRequest)
		return geo.Bbox{}, err
	}
	return bbox, nil
}

func (this *Handler) tileParams(w http.ResponseWriter, req *http.Request) (string, geo.Bbox, error) {
	callback := req.FormValue("callback")
	bbox, err := this.bboxFormValue(w, req)
	if err != nil {
		return "", geo.Bbox{}, err
	}

	return callback, bbox, nil
}

func (this *Handler) tileParamsZ(w http.ResponseWriter, req *http.Request) (string, geo.Bbox, int, error) {

	callback := req.FormValue("callback")
	bbox, err := this.bboxFormValue(w, req)
	if err != nil {
		return "", geo.Bbox{}, 0, err
	}

	zoomStr := req.FormValue("zoom")
	zoom, err := strconv.Atoi(zoomStr)
	if err != nil {
		this.onError(w, err, fmt.Sprintf("Can not parse zoom value: %s", zoomStr), http.StatusBadRequest)
		return "", geo.Bbox{}, 0, err
	}

	return callback, bbox, zoom, nil
}

func (this *Handler) onError(w http.ResponseWriter, err error, msg string, statusCode int) {
	errStr := fmt.Sprintf("%s: %v", msg, err)
	log.Errorf(errStr)
	http.Error(w, errStr, statusCode)
}

func (this *Handler) onError500(w http.ResponseWriter, err error, msg string) {
	this.onError(w, err, msg, http.StatusInternalServerError)
}

func (this *Handler) CloseAndRemove(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}

func (this *Handler) geoParser(r io.ReadSeeker) (GeoParser, error) {
	gpxParser, err := InitGpxParser(r)
	if err == nil {
		return gpxParser, nil
	}
	log.Warn(err)
	r.Seek(0, 0)
	kmlParser, err := InitKmlParser(r)
	if err == nil {
		return kmlParser, nil
	}
	log.Warn(err)
	return nil, errors.New("Can not find valid parser for this format!")
}