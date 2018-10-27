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
	"strconv"
	"github.com/and-hom/wwmap/backend/clustering"
	"github.com/and-hom/wwmap/backend/ymaps"
)

type WhiteWaterHandler struct {
	App
	ResourceBase string
	ClusterMaker clustering.ClusterMaker
}

const PREVIEWS_COUNT int = 20

func (this *WhiteWaterHandler) Init() {
	this.Register("/ymaps-tile-ww", HandlerFunctions{Get:this.TileWhiteWaterHandler})
	this.Register("/whitewater", HandlerFunctions{Post: this.InsertWhiteWaterPoints, Put:this.InsertWhiteWaterPoints})
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

	token := req.FormValue("token")
	allowed := false
	if token != "" {
		allowed, err = this.CheckRoleAllowed(req, ADMIN, EDITOR)
		if err != nil {
			OnError500(w, err, "Can not get user info for token")
			return
		}
	}

	rivers, err := this.TileDao.ListRiversWithBounds(bbox, allowed, PREVIEWS_COUNT)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not read whitewater points for bbox %s", bbox.String()))
		return
	}

	features, err := ymaps.WhiteWaterPointsToYmaps(this.ClusterMaker, rivers, bbox, zoom,
		this.ResourceBase, skip, this.processForWeb, getLinkMaker(req.FormValue("link_type")))
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not cluster: %s", bbox.String()))
		return
	}
	featureCollection := MkFeatureCollection(features)

	w.Write(this.JsonpAnswer(callback, featureCollection, "{}"))
}

func getLinkMaker(linkType string) ymaps.LinkMaker {
	switch linkType {
	case "none":
		return ymaps.NoneLinkMaker{}
	case "wwmap":
		return ymaps.WwmapLinkMaker{}
	case "huskytm":
		return ymaps.HuskytmLinkMaker{}
	default:
		return ymaps.FromSpotLinkMaker{}
	}
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