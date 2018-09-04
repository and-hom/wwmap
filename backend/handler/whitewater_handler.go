package handler

import (
	"net/http"
	"encoding/json"
	"fmt"
	"io/ioutil"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
	"math"
	"strconv"
	"github.com/gorilla/mux"
	"github.com/and-hom/wwmap/backend/clustering"
	"github.com/and-hom/wwmap/backend/ymaps"
)

type WhiteWaterHandler struct {
	App
	ResourceBase string
	ClusterMaker clustering.ClusterMaker
}

func (this *WhiteWaterHandler) Init(r *mux.Router) {
	this.Register(r, "/ymaps-tile-ww", HandlerFunctions{Get:this.TileWhiteWaterHandler})
	this.Register(r, "/whitewater", HandlerFunctions{Post: this.InsertWhiteWaterPoints, Put:this.InsertWhiteWaterPoints})
}

func (this *WhiteWaterHandler) TileWhiteWaterHandler(w http.ResponseWriter, req *http.Request) {
	this.collectReferer(req)

	callback, bbox, zoom, err := this.tileParamsZ(w, req)
	if err != nil {
		return
	}

	skipIdStr := req.FormValue("skip")
	skip := int64(0)
	if skipIdStr != "" {
		skip, err = strconv.ParseInt(skipIdStr, 10, 64)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not parse skip id %s", skipIdStr))
			return
		}
	}

	rivers, err := this.RiverDao.ListRiversWithBounds(bbox, math.MaxInt32)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not read whitewater points for bbox %s", bbox.String()))
		return
	}

	features, err := ymaps.WhiteWaterPointsToYmaps(this.ClusterMaker, rivers, bbox, zoom, this.ResourceBase, skip)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not cluster: %s", bbox.String()))
		return
	}
	featureCollection := MkFeatureCollection(features)

	w.Write(this.JsonpAnswer(callback, featureCollection, "{}"))
}

func (this *WhiteWaterHandler) InsertWhiteWaterPoints(w http.ResponseWriter, r *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, r, ADMIN, EDITOR) {
		return
	}

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

	err = this.WhiteWaterDao.InsertWhiteWaterPoints(wwPoints...)

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