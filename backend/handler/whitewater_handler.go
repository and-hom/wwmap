package handler

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/backend/clustering"
	"github.com/and-hom/wwmap/backend/ymaps"
	"github.com/and-hom/wwmap/lib/dao"
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
	this.Register("/search", HandlerFunctions{Post: this.search})
}

func (this *WhiteWaterHandler) TileWhiteWaterHandler(w http.ResponseWriter, req *http.Request) {
	this.collectReferer(req)
	w.Header().Set("Content-Type", "application/javascript")

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

	riverIdStr := req.FormValue("river")
	riverId := int64(0)
	if riverIdStr != "" {
		riverId, err = strconv.ParseInt(riverIdStr, 10, 64)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not parse river id %s", skipIdStr))
			return
		}
	}

	sessionId := req.FormValue("session_id")
	allowed := false
	if sessionId != "" {
		_, allowed, err = this.CheckRoleAllowed(req, dao.ADMIN, dao.EDITOR)
		if err != nil {
			OnError500(w, err, "Can not get user info for token")
			return
		}
	}

	var features []Feature

	if onlyId != 0 {
		spot, err := this.WhiteWaterDao.Find(onlyId)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not find whitewater point for id %d", onlyId))
			return
		}
		sp := dao.Spot{
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
		rws := dao.RiverWithSpots{
			IdTitle:   river.IdTitle,
			CountryId: river.Region.CountryId,
			RegionId:  river.Region.Id,
			Spots:     []dao.Spot{},
		}

		features, err = ymaps.SingleWhiteWaterPointToYmaps(sp, rws, this.ResourceBase, this.processForWeb, getLinkMaker(req.FormValue("link_type")))
	} else if riverId != 0 {
		river, found, err := this.TileDao.GetRiverById(riverId, PREVIEWS_COUNT)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not read whitewater points for river id %d", riverId))
			return
		}
		if !found {
			OnErrorWithCustomLogging(w, nil, fmt.Sprintf("Spots for river %d not found", riverId), http.StatusNotFound, func(s string) {
				logrus.Debug(s)
			})
			return
		}

		features = ymaps.WhiteWaterPointsToYmapsNoCluster([]dao.RiverWithSpots{river},
			this.ResourceBase, skip, this.processForWeb, getLinkMaker(req.FormValue("link_type")))
	} else {
		rivers, err := this.TileDao.ListRiversWithBounds(bbox, PREVIEWS_COUNT, allowed)
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
		OnError500(w, err, "Can not select rivers")
		return
	}

	this.JsonAnswer(w, SearchResp{
		Spots:        spots,
		Rivers:       rivers,
		ResourceBase: this.ResourceBase,
	})
}

type SearchResp struct {
	Spots        []dao.WhiteWaterPointWithRiverTitle `json:"spots"`
	Rivers       []dao.RiverTitle                    `json:"rivers"`
	ResourceBase string                              `json:"resource_base"`
}
