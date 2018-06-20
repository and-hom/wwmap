package main

import (
	"net/http"
	. "github.com/and-hom/wwmap/lib/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"strconv"
	"fmt"
	"github.com/Sirupsen/logrus"
)

type GeoHierarchyHandler struct {
	Handler
}

func (this *GeoHierarchyHandler) ListCountries(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")

	countries, err := this.countryDao.List()
	if err != nil {
		OnError500(w, err, "Can not list countries")
	}
	bytes, err := json.Marshal(countries)
	if err != nil {
		OnError500(w, err, "Can not serialize countries")
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) ListRegions(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")

	pathParams := mux.Vars(r)
	countryId, err := strconv.ParseInt(pathParams["countryId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	regions, err := this.regionDao.List(countryId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list regions of country %d", countryId))
	}
	bytes, err := json.Marshal(regions)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize regions of country %d", countryId))
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) ListCountryRivers(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")

	pathParams := mux.Vars(r)
	countryId, err := strconv.ParseInt(pathParams["countryId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	rivers, err := this.riverDao.ListByCountry(countryId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list rivers of country %d", countryId))
	}
	bytes, err := json.Marshal(rivers)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize rivers of country %d", countryId))
	}

	w.Write(bytes)
}

func (this *GeoHierarchyHandler) ListRegionRivers(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")

	pathParams := mux.Vars(r)
	regionId, err := strconv.ParseInt(pathParams["regionId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	rivers, err := this.riverDao.ListByRegion(regionId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list rivers of region %d", regionId))
	}
	bytes, err := json.Marshal(rivers)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize rivers of region %d", regionId))
	}

	w.Write(bytes)
}

const DEFAULT_REPORT_GROUP_LIMIT int = 20

func (this *GeoHierarchyHandler) ListRiverReports(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")

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
			logrus.Warn("Can not parse limit-by-group parameter: ", limitByGroupStr, err)
			groupLimit = DEFAULT_REPORT_GROUP_LIMIT
		}
	}

	voyageReports, err := this.voyageReportDao.List(riverId, int(groupLimit))
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list reports of river %d", riverId))
	}
	bytes, err := json.Marshal(voyageReports)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize reports of river %d", riverId))
	}

	w.Write(bytes)
}


func (this *GeoHierarchyHandler) ListSpots(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET, OPTIONS, POST, DELETE")

	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	voyageReports, err := this.whiteWaterDao.ListByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list spots of river %d", riverId))
	}
	bytes, err := json.Marshal(voyageReports)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize spots of river %d", riverId))
	}

	w.Write(bytes)
}
