package main

import (
	"net/http"
	"fmt"
	"strconv"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
)

type RiverHandler struct {
	Handler
}

func (this *RiverHandler) GetNearestRivers(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET")
	lat_s := r.FormValue("lat")
	lat, err := strconv.ParseFloat(lat_s, 64)
	if err != nil {
		OnError(w, err, fmt.Sprintf("Can not parse lat parameter: %s", lat_s), 400)
		return
	}
	lon_s := r.FormValue("lon")
	lon, err := strconv.ParseFloat(lon_s, 64)
	if err != nil {
		OnError(w, err, fmt.Sprintf("Can not parse lon parameter: %s", lon_s), 400)
		return
	}
	point := Point{Lat:lat, Lon:lon}
	rivers, err := this.riverDao.NearestRivers(point, 5)
	if err != nil {
		OnError500(w, err, "Can not select rivers")
		return
	}
	w.Write([]byte(this.JsonStr(rivers, "[]")))
}

func (this *RiverHandler) GetVisibleRivers(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, "GET")

	bbox, err := this.bboxFormValue(w, req)
	if err != nil {
		return
	}

	rivers, err := this.riverDao.ListRiversWithBounds(bbox, 30)
	if err != nil {
		OnError500(w, err, "Can not select rivers")
		return
	}

	for i := 0; i < len(rivers); i++ {
		river := &rivers[i]
		river.Bounds = river.Bounds.WithMargins(0.05)
	}
	w.Write([]byte(this.JsonStr(rivers, "[]")))
}
