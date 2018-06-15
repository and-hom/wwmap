package main

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"encoding/json"
	"fmt"
	"io/ioutil"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
)

const PREVIEWS_COUNT int = 20

type WhiteWaterHandler struct {
	Handler
	resourceBase string
}

func (this *WhiteWaterHandler) TileWhiteWaterHandler(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, "GET, OPTIONS")

	callback, bbox, zoom, err := this.tileParamsZ(w, req)
	if err != nil {
		return
	}

	points, err := this.whiteWaterDao.ListByBbox(bbox)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not read whitewater points for bbox %s", bbox.String()))
		return
	}
	for i := 0; i < len(points); i++ {
		imgs, err := this.imgDao.List(points[i].Id, PREVIEWS_COUNT)
		if err != nil {
			log.Warnf("Can not read whitewater point images for point %d: %s", points[i].Id, err.Error())
			continue
		}
		points[i].Images = imgs
	}

	featureCollection := MkFeatureCollection(whiteWaterPointsToYmaps(this.clusterMaker, points, bbox.Width(), bbox.Height(), zoom, this.resourceBase))
	log.Infof("Found %d", len(featureCollection.Features))

	w.Write(this.JsonpAnswer(callback, featureCollection, "{}"))
}

func (this *WhiteWaterHandler) AddWhiteWaterPoints(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")
	err := r.ParseForm()
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}

	wwPoints, err := this.parseWhiteWaterPointsForm(w, r)
	if err != nil {
		OnError500(w, err, "Can not read request")
		return
	}

	err = this.whiteWaterDao.AddWhiteWaterPoints(wwPoints...)
	fmt.Printf("%v\n", wwPoints)

	if err != nil {
		OnError500(w, err, "Can not insert")
		return
	}
}

func (this *WhiteWaterHandler) parseWhiteWaterPointsForm(w http.ResponseWriter, r *http.Request) ([]WhiteWaterPoint, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []WhiteWaterPoint{}, err
	}
	var points []WhiteWaterPoint
	err = json.Unmarshal(body, &points)
	if err != nil {
		return []WhiteWaterPoint{}, err
	}
	return points, nil
}