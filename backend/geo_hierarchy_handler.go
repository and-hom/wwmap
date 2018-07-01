package main

import (
	"net/http"
	. "github.com/and-hom/wwmap/lib/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"strconv"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"io/ioutil"
	"github.com/and-hom/wwmap/lib/util"
)

type GeoHierarchyHandler struct {
	Handler
	regions map[int64]dao.Region
}

type RiverDto struct {
	Id      int64 `json:"id"`
	Title   string `json:"title"`
	Aliases []string `json:"aliases"`
	Region  dao.Region `json:"region"`
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
		region, err := this.regionDao.Get(id)
		if err != nil {
			log.Errorf("Can not get region by id :", id, err)
			return dao.Region{Id:0, CountryId:0, Title:"-"}
		}
		this.regions[id] = region
		return region
	}
}

func (this *GeoHierarchyHandler) ListCountries(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)

	countries, err := this.countryDao.List()
	if err != nil {
		OnError500(w, err, "Can not list countries")
		return
	}
	bytes, err := json.Marshal(countries)
	if err != nil {
		OnError500(w, err, "Can not serialize countries")
		return
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) ListRegions(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)

	pathParams := mux.Vars(r)
	countryId, err := strconv.ParseInt(pathParams["countryId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	regions, err := this.regionDao.List(countryId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list regions of country %d", countryId))
		return
	}
	bytes, err := json.Marshal(regions)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize regions of country %d", countryId))
		return
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) ListAllRegions(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)

	regions, err := this.regionDao.ListAllWithCountry()
	if err != nil {
		OnError500(w, err, "Can not list regions")
		return
	}
	bytes, err := json.Marshal(regions)
	if err != nil {
		OnError500(w, err, "Can not serialize regions")
		return
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) ListCountryRivers(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)

	pathParams := mux.Vars(r)
	countryId, err := strconv.ParseInt(pathParams["countryId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	rivers, err := this.riverDao.ListByCountry(countryId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list rivers of country %d", countryId))
		return
	}
	bytes, err := json.Marshal(rivers)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize rivers of country %d", countryId))
		return
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) ListRegionRivers(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)

	pathParams := mux.Vars(r)
	regionId, err := strconv.ParseInt(pathParams["regionId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	rivers, err := this.riverDao.ListByRegion(regionId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list rivers of region %d", regionId))
		return
	}
	bytes, err := json.Marshal(rivers)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize rivers of region %d", regionId))
		return
	}

	w.Write(bytes)
}

const DEFAULT_REPORT_GROUP_LIMIT int = 20

func (this *GeoHierarchyHandler) ListRiverReports(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)

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

	voyageReports, err := this.voyageReportDao.List(riverId, int(groupLimit))
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list reports of river %d", riverId))
		return
	}
	bytes, err := json.Marshal(voyageReports)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize reports of river %d", riverId))
		return
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) ListSpots(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)

	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	voyageReports, err := this.whiteWaterDao.ListByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list spots of river %d", riverId))
		return
	}
	bytes, err := json.Marshal(voyageReports)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize spots of river %d", riverId))
		return
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) GetRiver(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)

	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.writeRiver(riverId, w)
}

func (this *GeoHierarchyHandler) SaveRiver(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN) {
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

	riverForDb := dao.RiverTitle{
		IdTitle:dao.IdTitle{
			Id:river.Id,
			Title:river.Title,
		},
		RegionId:river.Region.Id,
		Aliases:river.Aliases,
	}
	err = this.riverDao.Save(riverForDb)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not save river %d", string(bodyBytes)))
		return
	}

	this.writeRiver(river.Id, w)
}

func (this *GeoHierarchyHandler) RemoveRiver(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN) {
		return
	}

	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = this.riverDao.Remove(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not remove river by id: %d", riverId))
		return
	}
}

func (this *GeoHierarchyHandler) writeRiver(riverId int64, w http.ResponseWriter) {
	river, err := this.riverDao.Find(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get river %d", riverId))
		return
	}

	region, err := this.regionDao.Get(river.RegionId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get region for river %d", riverId))
		return
	}

	riverWithRegion := RiverDto{
		Id:river.Id,
		Title:river.Title,
		Aliases:river.Aliases,
		Region:region,
	}
	bytes, err := json.Marshal(riverWithRegion)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize river %d", riverId))
		return
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) FilterRivers(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS")
	JsonResponse(w)

	limit := 20

	query := util.FirstOr(r.URL.Query()["q"], "")

	rivers, err := this.riverDao.ListByFirstLetters(query, limit)
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
	bytes, err := json.Marshal(dtos)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize rivers: %v", dtos))
		return
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) GetSpot(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)

	pathParams := mux.Vars(r)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.writeSpot(spotId, w)
}

func (this *GeoHierarchyHandler) SaveSpot(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN) {
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

	err = this.whiteWaterDao.UpdateWhiteWaterPointsFull(spot)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not save spot %d", string(bodyBytes)))
		return
	}

	this.writeSpot(spot.Id, w)
}

func (this *GeoHierarchyHandler) RemoveSpot(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")
	JsonResponse(w)
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN) {
		return
	}

	pathParams := mux.Vars(r)
	spotId, err := strconv.ParseInt(pathParams["spotId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = this.whiteWaterDao.Remove(spotId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not remove spot by id: %d", spotId))
		return
	}
}

func (this *GeoHierarchyHandler) writeSpot(spotId int64, w http.ResponseWriter) {
	spot, err := this.whiteWaterDao.FindFull(spotId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get spot %d", spotId))
		return
	}

	bytes, err := json.Marshal(spot)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize spot %d", spotId))
		return
	}

	w.Write(bytes)
}