package handler

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/model"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/ptrv/go-gpx"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type GeoHierarchyHandler struct {
	App
	ImgStorage               blob.BlobStorage
	PreviewImgStorage        blob.BlobStorage
	RiverPassportPdfStorage  blob.BlobStorage
	RiverPassportHtmlStorage blob.BlobStorage
	regions                  map[int64]dao.Region
}

func (this *GeoHierarchyHandler) Init() {
	this.Register("/country", HandlerFunctions{Get: this.ListCountries})
	this.Register("/country/{countryId}/region", HandlerFunctions{Get: this.ListRegions})
	this.Register("/country/{countryId}/region/{regionId}/river", HandlerFunctions{Get: this.ListRegionRivers})
	this.Register("/country/{countryId}/river", HandlerFunctions{Get: this.ListCountryRivers})

	this.Register("/region", HandlerFunctions{Get: this.ListAllRegions})
	this.Register("/region/{regionId}", HandlerFunctions{Get: this.GetRegion})

	this.Register("/river", HandlerFunctions{Get: this.FilterRivers})
	this.Register("/river/{riverId}", HandlerFunctions{Get: this.GetRiver,
		Put:    this.ForRoles(this.SaveRiver, dao.ADMIN, dao.EDITOR),
		Post:   this.ForRoles(this.SaveRiver, dao.ADMIN, dao.EDITOR),
		Delete: this.ForRoles(this.RemoveRiver, dao.ADMIN, dao.EDITOR)})
	this.Register("/river/{riverId}/reports", HandlerFunctions{Get: this.ListRiverReports})
	this.Register("/river/{riverId}/spots", HandlerFunctions{Get: this.ListSpots})
	this.Register("/river/{riverId}/center", HandlerFunctions{Get: this.GetRiverCenter})
	this.Register("/river/{riverId}/bounds", HandlerFunctions{Get: this.GetRiverBounds})
	this.Register("/river/{riverId}/gpx", HandlerFunctions{Post: this.ForRoles(this.UploadGpx, dao.ADMIN, dao.EDITOR),
		Put: this.ForRoles(this.UploadGpx, dao.ADMIN, dao.EDITOR)})
	this.Register("/river/{riverId}/pdf", HandlerFunctions{Get: this.GetRiverPassportPdf})
	this.Register("/river/{riverId}/html", HandlerFunctions{Get: this.GetRiverPassportHtml})
	this.Register("/river/{riverId}/visible", HandlerFunctions{
		Post: this.ForRoles(this.SetRiverVisible, dao.ADMIN, dao.EDITOR),
		Put:  this.ForRoles(this.SetRiverVisible, dao.ADMIN, dao.EDITOR)})
	this.Register("/spot/{spotId}", HandlerFunctions{Get: this.GetSpot,
		Post:   this.ForRoles(this.SaveSpot, dao.ADMIN, dao.EDITOR),
		Put:    this.ForRoles(this.SaveSpot, dao.ADMIN, dao.EDITOR),
		Delete: this.ForRoles(this.RemoveSpot, dao.ADMIN, dao.EDITOR)})
}

type RiverDto struct {
	Id          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Aliases     []string               `json:"aliases"`
	Region      dao.Region             `json:"region"`
	Description string                 `json:"description,omitempty"`
	Visible     bool                   `json:"visible"`
	Props       map[string]interface{} `json:"props"`
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

func (this *GeoHierarchyHandler) GetRiverBounds(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	bounds, err := this.WhiteWaterDao.GetRiverBounds(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get bounds of river %d", riverId))
		return
	}
	this.JsonAnswer(w, bounds)
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
		spot.River = dao.RiverWithRegion{IdTitle: dao.IdTitle{Id: riverId}}
		spot.Point = geo.PointOrLine{Point: &geo.Point{Lat: wpt.Lat, Lon: wpt.Lon}}
		spot.ShortDesc = wpt.Desc
		spot.Category = model.SportCategory{Category: model.UNDEFINED_CATEGORY}
		spot.Aliases = []string{}
		_, err = this.WhiteWaterDao.InsertWhiteWaterPointFull(spot)
		if err != nil {
			OnError500(w, err, "Can not insert spot")
			return
		}
	}

	this.LogUserEvent(req, RIVER_LOG_ENTRY_TYPE, riverId, dao.ENTRY_TYPE_MODIFY, "Upload GPX")
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
	river := RiverDto{}
	body, err := decodeJsonBody(r, &river)
	if err != nil {
		OnError500(w, err, "Can not parse json from request body: "+body)
		return
	}

	if len(strings.TrimSpace(river.Title)) == 0 {
		OnError(w, errors.New(""), "Can not save river with empty name", http.StatusBadRequest)
		return
	}

	regionId := river.Region.Id
	if regionId == 0 {
		log.Error(river.Region.CountryId)
		fakeRegion, found, err := this.RegionDao.GetFake(river.Region.CountryId)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not get fake region for country: %d", river.Region.CountryId))
			return
		}
		if found {
			regionId = fakeRegion.Id
		} else {
			regionId, err = this.RegionDao.CreateFake(river.Region.CountryId)
			if err != nil {
				OnError500(w, err, fmt.Sprintf("Can not create fake region for country: %d", river.Region.CountryId))
				return
			}
			log.Errorf("RegionId = %v", regionId)
		}
	}

	riverForDb := dao.River{
		RiverTitle: dao.RiverTitle{
			IdTitle: dao.IdTitle{
				Id:    river.Id,
				Title: river.Title,
			},
			Region:  dao.Region{Id: regionId},
			Aliases: river.Aliases,
			Props:   river.Props,
		},
		Description: river.Description,
	}

	var id int64
	var logEntryType dao.ChangesLogEntryType
	if river.Id > 0 {
		err = this.RiverDao.SaveFull(riverForDb)
		id = river.Id
		logEntryType = dao.ENTRY_TYPE_MODIFY
	} else {
		id, err = this.RiverDao.Insert(riverForDb)
		logEntryType = dao.ENTRY_TYPE_CREATE
	}
	if err != nil {
		OnError500(w, err, "Can not save river: "+body)
		return
	}

	this.writeRiver(id, w)
	this.LogUserEvent(r, RIVER_LOG_ENTRY_TYPE, id, logEntryType, riverForDb.Title)
}

func (this *GeoHierarchyHandler) SetRiverVisible(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	visible := false
	body, err := decodeJsonBody(r, &visible)
	if err != nil {
		OnError(w, err, "Failed to parse request body: "+body, http.StatusBadRequest)
		return
	}

	err = this.RiverDao.SetVisible(riverId, visible)
	if err != nil {
		OnError500(w, err, "Can not update river")
		return
	}

	this.writeRiver(riverId, w)

	this.LogUserEvent(r, RIVER_LOG_ENTRY_TYPE, riverId, dao.ENTRY_TYPE_MODIFY, fmt.Sprintf("visible=%t", visible))
}

func (this *GeoHierarchyHandler) RemoveRiver(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	imgs, err := this.ImgDao.ListAllByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get images for river: %d %v", riverId, err))
		return
	}
	this.removeImageData(r, imgs)

	err = this.WaterWayDao.UnlinkRiver(riverId, nil)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not unlink river and waterway: %d %v", riverId, err))
		return
	}

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
	this.LogUserEvent(r, RIVER_LOG_ENTRY_TYPE, riverId, dao.ENTRY_TYPE_DELETE, "")
}

func (this *GeoHierarchyHandler) writeRiver(riverId int64, w http.ResponseWriter) {
	river, err := this.RiverDao.Find(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get river %d", riverId))
		return
	}

	riverWithRegion := RiverDto{
		Id:          river.Id,
		Title:       river.Title,
		Aliases:     river.Aliases,
		Region:      river.Region,
		Description: river.Description,
		Visible:     river.Visible,
		Props:       river.Props,
	}
	this.JsonAnswer(w, riverWithRegion)
}

func (this *GeoHierarchyHandler) FilterRivers(w http.ResponseWriter, r *http.Request) {
	limit := 20

	query := util.FirstOr(r.URL.Query()["q"], "")

	rivers, err := this.RiverDao.ListByFirstLetters(query, limit)
	if err != nil {
		OnError500(w, err, "Can not fetch rivers for query "+query)
		return
	}

	dtos := make([]RiverDto, len(rivers))
	for i := 0; i < len(rivers); i++ {
		river := &(rivers[i])
		dtos[i] = RiverDto{
			Id:      river.Id,
			Title:   river.Title,
			Aliases: river.Aliases,
			Region:  river.Region,
			Props:   river.Props,
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
	spot := dao.WhiteWaterPointFull{}
	body, err := decodeJsonBody(r, &spot)
	if err != nil {
		OnError500(w, err, "Can not parse json from request body: "+body)
		return
	}

	if spot.River.Id <= 0 {
		OnError(w, errors.New(""), "Can not save spot without river", http.StatusBadRequest)
		return
	}

	if len(strings.TrimSpace(spot.Title)) == 0 {
		OnError(w, errors.New(""), "Can not save spot with empty name", http.StatusBadRequest)
		return
	}

	var id int64
	var logType dao.ChangesLogEntryType
	if spot.Id > 0 {
		err = this.WhiteWaterDao.UpdateWhiteWaterPointsFull(spot)
		id = spot.Id
		logType = dao.ENTRY_TYPE_MODIFY
	} else {
		id, err = this.WhiteWaterDao.InsertWhiteWaterPointFull(spot)
		logType = dao.ENTRY_TYPE_CREATE
	}
	if err != nil {
		OnError500(w, err, "Can not save spot: "+body)
		return
	}

	this.writeSpot(id, w)

	this.LogUserEvent(r, SPOT_LOG_ENTRY_TYPE, id, logType, spot.Title)
}

func (this *GeoHierarchyHandler) RemoveSpot(w http.ResponseWriter, r *http.Request) {
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
	this.removeImageData(r, imgs)

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

	this.LogUserEvent(r, SPOT_LOG_ENTRY_TYPE, spotId, dao.ENTRY_TYPE_DELETE, "")
}

func (this *GeoHierarchyHandler) removeImageData(req *http.Request, imgs []dao.Img) {
	for _, img := range imgs {
		imgIdStr := fmt.Sprintf("%d", img.Id)

		err := this.ImgStorage.Remove(storageKeyById(imgIdStr))
		if err != nil {
			log.Errorf("Can not remove image for by id %d: %v", img.Id, err)
		}

		err = this.PreviewImgStorage.Remove(storageKeyById(imgIdStr))
		if err != nil {
			log.Errorf("Can not remove preview for by id %d: %v", img.Id, err)
		}
		this.LogUserEvent(req, RIVER_LOG_ENTRY_TYPE, img.Id, dao.ENTRY_TYPE_DELETE, "Recursively")
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

func (this *GeoHierarchyHandler) GetRiverPassportPdf(w http.ResponseWriter, req *http.Request) {
	this.getRiverPassport(w, req, "application/pdf", this.RiverPassportPdfStorage, ".pdf")
}

func (this *GeoHierarchyHandler) GetRiverPassportHtml(w http.ResponseWriter, req *http.Request) {
	this.getRiverPassport(w, req, "text/html", this.RiverPassportHtmlStorage, ".htm")
}

func (this *GeoHierarchyHandler) getRiverPassport(w http.ResponseWriter, req *http.Request,
	contentType string, storage blob.BlobStorage, suffix string) {
	pathParams := mux.Vars(req)
	w.Header().Set("Content-Type", contentType)
	length, err := storage.Length(pathParams["riverId"] + suffix)
	if err != nil {
		log.Warnf("Can not get river passport content length %s: %v", pathParams["riverId"], err)
		length = 0
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", length))
	r, err := storage.Read(pathParams["riverId"] + suffix)
	if err != nil {
		OnError500(w, err, "Can not get river passport")
		return
	}
	defer r.Close()
	io.Copy(w, r)

}
