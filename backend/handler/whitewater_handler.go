package handler

import (
	"fmt"
	"github.com/and-hom/wwmap/backend/clustering"
	"github.com/and-hom/wwmap/backend/handler/params"
	"github.com/and-hom/wwmap/backend/handler/tile_data_fetcher"
	"github.com/and-hom/wwmap/backend/handler/toggles"
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

func (this *WhiteWaterHandler) Init() {
	this.Register("/ymaps-tile-ww", HandlerFunctions{Get: this.TileWhiteWaterHandler})
	this.Register("/router-data", HandlerFunctions{Get: this.RouterData})
	this.Register("/river-path-segments", HandlerFunctions{Get: this.RiverPathSegments})
	this.Register("/search", HandlerFunctions{Post: this.search})
}

func (this *WhiteWaterHandler) TileWhiteWaterHandler(w http.ResponseWriter, req *http.Request) {
	this.collectReferer(req)
	w.Header().Set("Content-Type", "application/javascript")

	callback, bbox, zoom, err := TileParamsZ(w, req)
	if err != nil {
		OnError500(w, err, "Can't parse bbox")
		return
	}

	requestParams, req, err := params.Parse(req)
	if err != nil {
		OnError500(w, err, "Can't parse request params")
		return
	}

	featureToggles := toggles.ParseFeatureTogglesOrFallback(req, this.UserDao)

	var tdf tile_data_fetcher.TileDataFetcher

	if requestParams.Only != 0 {
		tdf = tile_data_fetcher.Spot(this.WhiteWaterDao, this.RiverDao,
			getLinkMaker(req.FormValue("link_type")), this.processForWeb, this.ResourceBase)
	} else if requestParams.River != 0 {
		tdf = tile_data_fetcher.River(this.TileDao, this.CampDao, this.ClusterMaker,
			getLinkMaker(req.FormValue("link_type")), this.processForWeb, this.ResourceBase)
	} else if requestParams.Region != 0 {
		tdf = tile_data_fetcher.Region(this.TileDao, this.ClusterMaker,
			getLinkMaker(req.FormValue("link_type")), this.processForWeb, this.ResourceBase)
	} else if requestParams.Country != 0 {
		tdf = tile_data_fetcher.Country(this.TileDao, this.ClusterMaker,
			getLinkMaker(req.FormValue("link_type")), this.processForWeb, this.ResourceBase)
	} else {
		tdf = tile_data_fetcher.World(this.TileDao, this.WaterWayDao, this.CampDao, this.ClusterMaker,
			getLinkMaker(req.FormValue("link_type")), this.processForWeb, this.ResourceBase)
	}

	features, req, err := tdf.Fetch(req, bbox, zoom, requestParams, featureToggles)
	if err != nil {
		switch e := err.(type) {
		case *tile_data_fetcher.DataFetchError:
			OnError(w, e.Cause(), e.Error(), e.HttpStatus())
		default:
			OnError500(w, err, "")
		}
		return
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
	_, bbox, _, err := TileParamsZ(w, req)
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
		showUnpublished,
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
		showUnpublished,
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
