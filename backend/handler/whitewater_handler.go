package handler

import (
	"fmt"
	"github.com/and-hom/wwmap/backend/clustering"
	"github.com/and-hom/wwmap/backend/handler/toggles"
	"github.com/and-hom/wwmap/backend/ymaps"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	log "github.com/sirupsen/logrus"
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
const MIN_ZOOM_SHOW_CAMPS int = 12
const MIN_ZOOM_SHOW_SLOPE int = 7

func (this *WhiteWaterHandler) Init() {
	this.Register("/ymaps-tile-ww", HandlerFunctions{Get: this.TileWhiteWaterHandler})
	this.Register("/router-data", HandlerFunctions{Get: this.RouterData})
	this.Register("/river-path-segments", HandlerFunctions{Get: this.RiverPathSegments})
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
			OnError500(w, err, fmt.Sprintf("Can not parse only id %s", onlyIdStr))
			return
		}
	}

	riverIdStr := req.FormValue("river")
	riverId := int64(0)
	if riverIdStr != "" {
		riverId, err = strconv.ParseInt(riverIdStr, 10, 64)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not parse river id %s", riverIdStr))
			return
		}
	}

	regionIdStr := req.FormValue("region")
	regionId := int64(0)
	if regionIdStr != "" {
		regionId, err = strconv.ParseInt(regionIdStr, 10, 64)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not parse region id %s", regionIdStr))
			return
		}
	}

	countryIdStr := req.FormValue("country")
	countryId := int64(0)
	if countryIdStr != "" {
		countryId, err = strconv.ParseInt(countryIdStr, 10, 64)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not parse country id %s", countryIdStr))
			return
		}
	}

	sessionId := req.FormValue("session_id")
	allowed := false
	if sessionId != "" {
		req, allowed, err = CheckRoleAllowed(req, this.UserDao, dao.ADMIN, dao.EDITOR)
		if err != nil {
			OnError500(w, err, "Can not get user info for token")
			return
		}
	}

	// Deprecated
	req, showUnpublishedOld := ShowUnpublished(req, this.UserDao)
	// Deprecated
	showCampsOld := GetBoolParameter(req, "show_camps", false) && zoom >= MIN_ZOOM_SHOW_CAMPS
	// Deprecated
	showSlopeOld := GetBoolParameter(req, "show_slope", false)

	featureToggles := toggles.ParseFeatureTogglesOrFallback(req, this.UserDao)
	ctx := req.Context()
	showCamps, ctx := featureToggles.GetShowCamps(ctx)
	showUnpublished, ctx := featureToggles.GetShowUnpublished(ctx)
	showSlope, ctx := featureToggles.GetShowSlope(ctx)
	if req.Context() != ctx {
		req = req.WithContext(ctx)
	}

	var features []Feature

	if onlyId != 0 {
		spot, found, err := this.WhiteWaterDao.Find(onlyId)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not select whitewater point for id %d", onlyId))
			return
		}
		if !found {
			OnError(w, err, fmt.Sprintf("Can not find whitewater point for id %d", onlyId), http.StatusNotFound)
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
			Visible:   river.Visible,
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
				log.Debug(s)
			})
			return
		}

		features = ymaps.WhiteWaterPointsToYmapsNoCluster([]dao.RiverWithSpots{river},
			this.ResourceBase, skip, this.processForWeb, getLinkMaker(req.FormValue("link_type")))

		if showCamps || showCampsOld {
			camps, err := this.CampDao.FindWithinBoundsForRiver(bbox, riverId)
			if err != nil {
				log.Error(err)
			} else {
				features = append(features, ymaps.CampsToYmaps(camps, this.ResourceBase, skip)...)
			}
		}
	} else if regionId != 0 {
		rivers, err := this.TileDao.ListRiversInRegionWithBounds(bbox, regionId, PREVIEWS_COUNT, allowed)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not read whitewater points for river id %d", riverId))
			return
		}

		features, err = ymaps.WhiteWaterPointsToYmaps(this.ClusterMaker, rivers, bbox, zoom,
			this.ResourceBase, skip, this.processForWeb, getLinkMaker(req.FormValue("link_type")),
			[]dao.WaterWay{})
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not cluster: %s", bbox.String()))
			return
		}
	} else if countryId != 0 {
		rivers, err := this.TileDao.ListRiversInCountryWithBounds(bbox, countryId, PREVIEWS_COUNT, allowed)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not read whitewater points for river id %d", riverId))
			return
		}

		features, err = ymaps.WhiteWaterPointsToYmaps(this.ClusterMaker, rivers, bbox, zoom,
			this.ResourceBase, skip, this.processForWeb, getLinkMaker(req.FormValue("link_type")),
			[]dao.WaterWay{})
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not cluster: %s", bbox.String()))
			return
		}
	} else {
		rivers, err := this.TileDao.ListRiversWithBounds(
			bbox,
			PREVIEWS_COUNT,
			showUnpublished || showUnpublishedOld,
		)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not read whitewater points for bbox %s", bbox.String()))
			return
		}

		riverIds := make([]int64, len(rivers))
		for i := 0; i < len(rivers); i++ {
			riverIds[i] = rivers[i].Id
		}
		//paths, err := this.WaterWayDao.ListByRiverIds(riverIds...)
		//paths, err := this.WaterWayDao.ListByBbox(bbox)
		//if err != nil {
		//	OnError500(w, err, "aaaa")
		//	return
		//}

		paths := []dao.WaterWay{}

		features, err = ymaps.WhiteWaterPointsToYmaps(this.ClusterMaker, rivers, bbox, zoom,
			this.ResourceBase, skip, this.processForWeb, getLinkMaker(req.FormValue("link_type")),
			paths)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not cluster: %s", bbox.String()))
			return
		}

		if showCamps || showCampsOld {
			camps, err := this.CampDao.FindWithinBounds(bbox)
			if err != nil {
				log.Error(err)
			} else {
				features = append(features, ymaps.CampsToYmaps(camps, this.ResourceBase, skip)...)
			}
		}
		if showSlope || showSlopeOld && zoom >= MIN_ZOOM_SHOW_SLOPE {
			var experimentEnabled = false
			req, experimentEnabled, err = this.experimentalFeaturesEnabled(req)
			if err != nil {
				OnError500(w, err, fmt.Sprintf("Can not check experimental futures of current user"))
				return
			}
			if experimentEnabled {
				tracks, err := this.WaterWayDao.ListWithHeightsByBbox(bbox)
				if err != nil {
					log.Error(err)
				} else {
					features = append(features, ymaps.TracksToYmaps(tracks, OBJECT_TYPE_SLOPE)...)
				}
			}
		}
	}

	featureCollection := MkFeatureCollection(features)

	w.Write(JsonpAnswer(callback, featureCollection, "{}"))
}

func (this *WhiteWaterHandler) RiverPathSegments(w http.ResponseWriter, req *http.Request) {
	riverId, err := strconv.ParseInt(req.FormValue("riverId"), 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	ways, err := this.WaterWayDao.ListByRiverIdNonFlipped(riverId)
	if err != nil {
		OnError(w, err, "Can not select paths", http.StatusBadRequest)
		return
	}
	JsonAnswer(w, ways)
}

func (this *WhiteWaterHandler) RouterData(w http.ResponseWriter, req *http.Request) {
	_, bbox, _, err := this.tileParamsZ(w, req)
	ways, err := this.WaterWayDao.ListByBboxNonFilpped(bbox)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get waterways for bbox: %s", bbox.String()))
		return
	}
	waysMap := make(map[int64]dao.WaterWay4Router)
	for i := 0; i < len(ways); i++ {
		waysMap[ways[i].Id] = ways[i]
	}
	JsonAnswer(w, RouterDataResp{
		Tracks: waysMap,
	})
}

func (this *WhiteWaterHandler) RouterStaticData(w http.ResponseWriter, req *http.Request) {
	ids, err := this.WaterWayRefDao.RefsById()
	if err != nil {
		OnError500(w, err, "Can not get waterway ids")
		return
	}
	JsonAnswer(w, ids)
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

func (this *WhiteWaterHandler) search(w http.ResponseWriter, req *http.Request) {
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		OnError500(w, err, "Can not read body")
		return
	}

	featureToggles := toggles.ParseFeatureTogglesOrFallback(req, this.UserDao)
	req, showUnpublishedOld := ShowUnpublished(req, this.UserDao)

	showUnpublished, ctx := featureToggles.GetShowUnpublished(req.Context())
	if req.Context() != ctx {
		req = req.WithContext(ctx)
	}

	regionIdStr := req.FormValue("region")
	regionId := int64(0)
	if regionIdStr != "" {
		regionId, err = strconv.ParseInt(regionIdStr, 10, 64)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not parse region id %d", regionId))
			return
		}
	}

	countryIdStr := req.FormValue("country")
	countryId := int64(0)
	if countryIdStr != "" {
		countryId, err = strconv.ParseInt(countryIdStr, 10, 64)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not parse country id %d", countryId))
			return
		}
	}

	spots, err := this.WhiteWaterDao.FindByTitlePart(
		string(requestBody),
		regionId,
		countryId,
		30,
		0,
		showUnpublished || showUnpublishedOld,
	)
	if err != nil {
		OnError500(w, err, "Can not select spots")
		return
	}
	rivers, err := this.RiverDao.FindByTitlePart(
		string(requestBody),
		regionId, countryId,
		30,
		0,
		showUnpublished || showUnpublishedOld,
	)
	if err != nil {
		OnError500(w, err, "Can not select rivers")
		return
	}

	JsonAnswer(w, SearchResp{
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

type RouterDataResp struct {
	Tracks map[int64]dao.WaterWay4Router `json:"tracks"`
}
