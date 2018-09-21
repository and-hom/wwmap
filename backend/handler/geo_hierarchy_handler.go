package handler

import (
	"net/http"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
	"encoding/json"
	"github.com/gorilla/mux"
	"strconv"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"io/ioutil"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/pkg/errors"
	"github.com/ptrv/go-gpx"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/model"
	"github.com/and-hom/wwmap/lib/img_storage"
)

type GeoHierarchyHandler struct {
	App
	ImgStorage        img_storage.ImgStorage
	PreviewImgStorage img_storage.ImgStorage
	regions           map[int64]dao.Region
}

func (this *GeoHierarchyHandler) Init(r *mux.Router) {
	this.Register(r, "/country", HandlerFunctions{Get: this.ListCountries, })
	this.Register(r, "/country/{countryId}/region", HandlerFunctions{Get:this.ListRegions})
	this.Register(r, "/country/{countryId}/region/{regionId}/river", HandlerFunctions{Get:this.ListRegionRivers})
	this.Register(r, "/country/{countryId}/river", HandlerFunctions{Get:this.ListCountryRivers})

	this.Register(r, "/region", HandlerFunctions{Get:this.ListAllRegions})
	this.Register(r, "/region/{regionId}", HandlerFunctions{Get:this.GetRegion})

	this.Register(r, "/river", HandlerFunctions{Get:this.FilterRivers})
	this.Register(r, "/river/{riverId}", HandlerFunctions{Get:this.GetRiver, Put:this.SaveRiver, Post:this.SaveRiver, Delete:this.RemoveRiver})
	this.Register(r, "/river/{riverId}/reports", HandlerFunctions{Get:this.ListRiverReports})
	this.Register(r, "/river/{riverId}/spots", HandlerFunctions{Get:this.ListSpots})
	this.Register(r, "/river/{riverId}/center", HandlerFunctions{Get:this.GetRiverCenter})
	this.Register(r, "/river/{riverId}/gpx", HandlerFunctions{Post:this.UploadGpx, Put:this.UploadGpx})
	this.Register(r, "/river/{riverId}/visible", HandlerFunctions{Post:this.SetRiverVisible, Put:this.SetRiverVisible})

	this.Register(r, "/spot/{spotId}", HandlerFunctions{Get:this.GetSpot, Post:this.SaveSpot, Put:this.SaveSpot, Delete:this.RemoveSpot})
}

type RiverDto struct {
	Id          int64 `json:"id"`
	Title       string `json:"title"`
	Aliases     []string `json:"aliases"`
	Region      dao.Region `json:"region"`
	Description string `json:"description,omitempty"`
	Visible     bool `json:"visible"`
}

func (this *GeoHierarchyHandler) getRegion(id int64) dao.Region {
	if this.regions == nil {
		this.regions = make(map[int64]dao.Region)
	}

	region, found := this.regions[id]
	if found {
		return region
	} else {
		log.Debugf("Region id=%d not found in cache. Select.", id)
		region, err := this.RegionDao.Get(id)
		if err != nil {
			log.Errorf("Can not get region by id :", id, err)
			return dao.Region{Id:0, CountryId:0, Title:"-"}
		}
		this.regions[id] = region
		return region
	}
}

func (this *GeoHierarchyHandler) ListCountries(w http.ResponseWriter, r *http.Request) {
	countries, err := this.CountryDao.List()
	if err != nil {
		OnError500(w, err, "Can not list countries")
		return
	}
	this.JsonAnswer(w, countries)
}

func (this *GeoHierarchyHandler) ListRegions(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	countryId, err := strconv.ParseInt(pathParams["countryId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	regions, err := this.RegionDao.List(countryId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list regions of country %d", countryId))
		return
	}
	this.JsonAnswer(w, regions)
}

func (this *GeoHierarchyHandler) ListAllRegions(w http.ResponseWriter, r *http.Request) {
	regions, err := this.RegionDao.ListAllWithCountry()
	if err != nil {
		OnError500(w, err, "Can not list regions")
		return
	}
	this.JsonAnswer(w, regions)
}

func (this *GeoHierarchyHandler) GetRegion(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["regionId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.writeRegion(riverId, w)
}

func (this *GeoHierarchyHandler) writeRegion(regionId int64, w http.ResponseWriter) {
	region, err := this.RegionDao.Get(regionId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get region %d", regionId))
		return
	}
	this.JsonAnswer(w, region)
}

func (this *GeoHierarchyHandler) ListCountryRivers(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	countryId, err := strconv.ParseInt(pathParams["countryId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	rivers, err := this.RiverDao.ListByCountry(countryId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list rivers of country %d", countryId))
		return
	}
	this.JsonAnswer(w, rivers)
}

func (this *GeoHierarchyHandler) ListRegionRivers(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	regionId, err := strconv.ParseInt(pathParams["regionId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	rivers, err := this.RiverDao.ListByRegion(regionId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list rivers of region %d", regionId))
		return
	}
	this.JsonAnswer(w, rivers)
}

const DEFAULT_REPORT_GROUP_LIMIT int = 20

func (this *GeoHierarchyHandler) ListRiverReports(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	limitByGroupStr := r.FormValue("limit-by-group")
	groupLimit := DEFAULT_REPORT_GROUP_LIMIT
	if limitByGroupStr != "" {
		groupLimit64, err := strconv.ParseInt(limitByGroupStr, 10, 32)
		groupLimit = int(groupLimit64)
		if err != nil {
			log.Warn("Can not parse limit-by-group parameter: ", limitByGroupStr, err)
			groupLimit = DEFAULT_REPORT_GROUP_LIMIT
		}
	}

	voyageReports, err := this.VoyageReportDao.List(riverId, int(groupLimit))
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list reports of river %d", riverId))
		return
	}
	this.JsonAnswer(w, voyageReports)
}

func (this *GeoHierarchyHandler) ListSpots(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	voyageReports, err := this.WhiteWaterDao.ListByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list spots of river %d", riverId))
		return
	}
	this.JsonAnswer(w, voyageReports)
}

func (this *GeoHierarchyHandler) GetRiverCenter(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	centroid, err := this.WhiteWaterDao.GetGeomCenterByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get centroid of river %d", riverId))
		return
	}
	this.JsonAnswer(w, centroid)
}

func (this *GeoHierarchyHandler) UploadGpx(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = req.ParseMultipartForm(128 * 1024 * 1024)
	if err != nil {
		OnError500(w, err, "Can not parse multipart form")
		return
	}
	f, _, err := req.FormFile("file")
	if err != nil {
		OnError500(w, err, "Can not get uploaded file")
		return
	}
	defer f.Close()

	gpx_data, err := gpx.Parse(f)
	if err != nil {
		OnError(w, err, "Can not parse gpx", http.StatusBadRequest)
		return
	}

	for _, wpt := range gpx_data.Waypoints {
		spot := dao.WhiteWaterPointFull{}
		spot.Title = wpt.Name
		spot.River = dao.IdTitle{Id:riverId}
		spot.Point = geo.Point{Lat:wpt.Lat, Lon:wpt.Lon}
		spot.ShortDesc = wpt.Desc
		spot.Category = model.SportCategory{Category:model.UNDEFINED_CATEGORY}
		_, err = this.WhiteWaterDao.InsertWhiteWaterPointFull(spot)
		if err != nil {
			OnError500(w, err, "Can not insert spot")
			return
		}
	}
}

func (this *GeoHierarchyHandler) GetRiver(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.writeRiver(riverId, w)
}

func (this *GeoHierarchyHandler) SaveRiver(w http.ResponseWriter, r *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN, dao.EDITOR) {
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		OnError500(w, err, "Can not read request body")
		return
	}
	river := RiverDto{}
	err = json.Unmarshal(bodyBytes, &river)
	if err != nil {
		OnError500(w, err, "Can not parse json from request body: " + string(bodyBytes))
		return
	}

	riverForDb := dao.River{
		RiverTitle: dao.RiverTitle{
			IdTitle:dao.IdTitle{
				Id:river.Id,
				Title:river.Title,
			},
			RegionId:river.Region.Id,
			Aliases:river.Aliases,
		},
		Description:river.Description,
	}

	var id int64
	if river.Id > 0 {
		err = this.RiverDao.Save(riverForDb)
		id = river.Id
	} else {
		id, err = this.RiverDao.Insert(riverForDb)
	}
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not save river %s", string(bodyBytes)))
		return
	}

	this.writeRiver(id, w)
}

func (this *GeoHierarchyHandler) SetRiverVisible(w http.ResponseWriter, r *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN, dao.EDITOR) {
		return
	}

	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		OnError500(w, err, "Can not read body")
		return
	}
	visible := false
	json.Unmarshal(bodyBytes, &visible)

	this.RiverDao.SetVisible(riverId, visible)

	this.writeRiver(riverId, w)
}

func (this *GeoHierarchyHandler) RemoveRiver(w http.ResponseWriter, r *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN, dao.EDITOR) {
		return
	}

	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	imgs, err := this.ImgDao.ListAllByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get images for river: %d", riverId))
		return
	}
	this.removeImageData(imgs)

	err = this.Storage.WithinTx(func(tx interface{}) error {
		err = this.ImgDao.RemoveByRiver(riverId, tx)
		if err != nil {
			return err
		}
		err := this.WhiteWaterDao.RemoveByRiver(riverId, tx)
		if err != nil {
			return err
		}
		err = this.VoyageReportDao.RemoveRiverLink(riverId, tx)
		if err != nil {
			return err
		}
		return this.RiverDao.Remove(riverId, tx)
	})

	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not remove river by id: %d", riverId))
		return
	}
}

func (this *GeoHierarchyHandler) writeRiver(riverId int64, w http.ResponseWriter) {
	river, err := this.RiverDao.Find(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get river %d", riverId))
		return
	}

	region, err := this.RegionDao.Get(river.RegionId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get region for river %d", riverId))
		return
	}

	riverWithRegion := RiverDto{
		Id:river.Id,
		Title:river.Title,
		Aliases:river.Aliases,
		Region:region,
		Description:river.Description,
		Visible: river.Visible,
	}
	this.JsonAnswer(w, riverWithRegion)
}

func (this *GeoHierarchyHandler) FilterRivers(w http.ResponseWriter, r *http.Request) {
	limit := 20

	query := util.FirstOr(r.URL.Query()["q"], "")

	rivers, err := this.RiverDao.ListByFirstLetters(query, limit)
	if err != nil {
		OnError500(w, err, "Can not fetch rivers for query" + query)
		return
	}

	dtos := make([]RiverDto, len(rivers))
	for i := 0; i < len(rivers); i++ {
		river := &(rivers[i])
		dtos[i] = RiverDto{
			Id:river.Id,
			Title:river.Title,
			Aliases:river.Aliases,
			Region:this.getRegion(river.RegionId),
		}
	}
	this.JsonAnswer(w, dtos)
}

func (this *GeoHierarchyHandler) GetSpot(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.writeSpot(spotId, w)
}

func (this *GeoHierarchyHandler) SaveSpot(w http.ResponseWriter, r *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN, dao.EDITOR) {
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		OnError500(w, err, "Can not read request body")
		return
	}
	spot := dao.WhiteWaterPointFull{}
	err = json.Unmarshal(bodyBytes, &spot)
	if err != nil {
		OnError500(w, err, "Can not parse json from request body: " + string(bodyBytes))
		return
	}

	if spot.River.Id <= 0 {
		OnError(w, errors.New(""), "Can not save spot without river", http.StatusBadRequest)
		return
	}

	var id int64
	if spot.Id > 0 {
		err = this.WhiteWaterDao.UpdateWhiteWaterPointsFull(spot)
		id = spot.Id
	} else {
		id, err = this.WhiteWaterDao.InsertWhiteWaterPointFull(spot)
	}
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not save spot %d", string(bodyBytes)))
		return
	}

	this.writeSpot(id, w)
}

func (this *GeoHierarchyHandler) RemoveSpot(w http.ResponseWriter, r *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN, dao.EDITOR) {
		return
	}

	pathParams := mux.Vars(r)
	spotIdStr := pathParams["spotId"]
	spotId, err := strconv.ParseInt(spotIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	imgs, err := this.ImgDao.ListAllBySpot(spotId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list images for spot by id: %d", spotId))
		return
	}
	this.removeImageData(imgs)

	err = this.Storage.WithinTx(func(tx interface{}) error {
		err := this.ImgDao.RemoveBySpot(spotId, tx)
		if err != nil {
			return err
		}
		return this.WhiteWaterDao.Remove(spotId, tx)
	})

	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not remove spot by id: %d", spotId))
		return
	}
}

func (this *GeoHierarchyHandler)removeImageData(imgs []dao.Img) {
	for _, img := range imgs {
		imgIdStr := fmt.Sprintf("%d", img.Id)

		err := this.ImgStorage.Remove(imgIdStr)
		if err != nil {
			log.Errorf("Can not remove image for by id %d: %v", img.Id, err)
		}

		err = this.PreviewImgStorage.Remove(imgIdStr)
		if err != nil {
			log.Errorf("Can not remove preview for by id %d: %v", img.Id, err)
		}
	}
}

func (this *GeoHierarchyHandler) writeSpot(spotId int64, w http.ResponseWriter) {
	spot, err := this.WhiteWaterDao.FindFull(spotId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get spot %d", spotId))
		return
	}
	this.JsonAnswer(w, spot)
}