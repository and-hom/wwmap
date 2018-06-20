package main

import (
	"net/http"
	"encoding/json"
	"fmt"
	"io/ioutil"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
	"math"
)

const PREVIEWS_COUNT int = 20

type WhiteWaterHandler struct {
	Handler
	resourceBase  string
	clusterMaker ClusterMaker
}

func (this *WhiteWaterHandler) TileWhiteWaterHandler(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, "GET, OPTIONS")

	callback, bbox, zoom, err := this.tileParamsZ(w, req)
	if err != nil {
		return
	}

	rivers, err := this.riverDao.ListRiversWithBounds(bbox, math.MaxInt32)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not read whitewater points for bbox %s", bbox.String()))
		return
	}

	features, err := whiteWaterPointsToYmaps(this.clusterMaker, rivers, bbox, zoom, this.resourceBase)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not cluster: %s", bbox.String()))
		return
	}
	featureCollection := MkFeatureCollection(features)

	w.Write(this.JsonpAnswer(callback, featureCollection, "{}"))
}

func (this *WhiteWaterHandler) AddWhiteWaterPoints(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")
	found, err := this.CheckRoleAllowed(r, ADMIN)
	if err != nil {
		onPassportErr(err, w, "Can not do request to Yandex Passport")
		return
	}
	if !found {
		OnError(w, nil, "User not found", http.StatusUnauthorized)
		return 
	}


	err = r.ParseForm()
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