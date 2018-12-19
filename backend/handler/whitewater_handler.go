package handler

import (
	"encoding/json"
	"fmt"
	"github.com/and-hom/wwmap/backend/clustering"
	"github.com/and-hom/wwmap/backend/ymaps"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"io/ioutil"
	"net/http"
	"strconv"
)

type WhiteWaterHandler struct {
	App
	ResourceBase string
	ClusterMaker clustering.ClusterMaker
}

const PREVIEWS_COUNT int = 20

func (this *WhiteWaterHandler) Init() {
	this.Register("/ymaps-tile-ww", HandlerFunctions{Get: this.TileWhiteWaterHandler})
	this.Register("/whitewater", HandlerFunctions{Post: this.InsertWhiteWaterPoints, Put: this.InsertWhiteWaterPoints})
	this.Register("/search", HandlerFunctions{Post: this.search})
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

	onlyIdStr := req.FormValue("only")
	onlyId := int64(0)
	if onlyIdStr != "" {
		onlyId, err = strconv.ParseInt(onlyIdStr, 10, 64)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not parse only id %s", skipIdStr))
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

	var features []Feature

	if onlyId == 0 {
		rivers, err := this.TileDao.ListRiversWithBounds(bbox, allowed, PREVIEWS_COUNT)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not read whitewater points for bbox %s", bbox.String()))
			return
		}

		features, err = ymaps.WhiteWaterPointsToYmaps(this.ClusterMaker, rivers, bbox, zoom,
			this.ResourceBase, skip, this.processForWeb, getLinkMaker(req.FormValue("link_type")))
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not cluster: %s", bbox.String()))
			return
		}
	} else {
		spot, err := this.WhiteWaterDao.Find(onlyId)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not find whitewater point for id %d", onlyId))
			return
		}
		sp := Spot{
			IdTitle:     spot.IdTitle,
			Description: spot.ShortDesc,
			Link:        spot.Link,
			Point:       spot.Point,
			Images:      spot.Images,
			Category:    spot.Category,
		}

		river, err := this.RiverDao.Find(spot.RiverId)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not find river for id %d", spot.RiverId))
			return
		}
		rws := RiverWithSpots{
			IdTitle:   river.IdTitle,
			CountryId: river.Region.CountryId,
			RegionId:  river.Region.Id,
			Spots:     []Spot{},
		}

		features, err = ymaps.SingleWhiteWaterPointToYmaps(sp, rws, this.ResourceBase, this.processForWeb, getLinkMaker(req.FormValue("link_type")))
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

func (this *WhiteWaterHandler) search(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		OnError500(w, err, "Can not read body")
		return
	}

	spots, err := this.WhiteWaterDao.FindByTitlePart(string(requestBody), 30, 0)
	if err != nil {
		OnError500(w, err, "Can not select spots")
		return
	}
	rivers, err := this.RiverDao.FindByTitlePart(string(requestBody), 30, 0)
	if err != nil {
		OnError500(w, err, "Can not select spots")
		return
	}

	resp := SearchResp{
		Spots:        spots,
		Rivers:       rivers,
		ResourceBase: this.ResourceBase,
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		OnError500(w, err, "Can not marshal response")
		return
	}

	_, err = w.Write(respBytes)
	if err != nil {
		OnError500(w, err, "Can not write response")
		return
	}
}

type SearchResp struct {
	Spots        []WhiteWaterPointWithRiverTitle `json:"spots"`
	Rivers       []RiverTitle                    `json:"rivers"`
	ResourceBase string                          `json:"resource_base"`
}
