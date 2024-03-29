package handler

import (
	"fmt"
	"github.com/and-hom/wwmap/backend/handler/toggles"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type RiverHandler struct {
	App
	TransferDao                    dao.TransferDao
	ResourceBase                   string
	RiverPassportPdfStorageParams  config.BlobStorageParams
	RiverPassportHtmlStorageParams config.BlobStorageParams
	RiverPassportPdfStorage        blob.BlobStorage
	RiverPassportHtmlStorage       blob.BlobStorage
}

func (this *RiverHandler) Init() {
	this.Register("/visible-rivers", HandlerFunctions{Get: this.GetVisibleRivers})
	this.Register("/visible-rivers-lite", HandlerFunctions{Get: this.GetVisibleRiversLite})
	this.Register("/river-card/{riverId}", HandlerFunctions{Get: this.GetRiverCard})
}

const MAX_REPORTS_PER_SOURCE = 5
const RIVER_LIST_LIMIT = 30
const RIVER_BOUNDS_MARGINS_RATIO = 0.05

type VoyageReportDto struct {
	Id            int64  `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Year          int    `json:"year"`
	Url           string `json:"url"`
	SourceLogoUrl string `json:"source_logo_url"`
}

type VoyageReportListDto struct {
	Source  string            `json:"source"`
	Reports []VoyageReportDto `json:"reports"`
}

type RiverListDto struct {
	dao.RiverTitle
	Reports []VoyageReportDto `json:"reports"`
	PdfUrl  string            `json:"pdf"`
	HtmlUrl string            `json:"html"`
}

type RiverListLiteDto struct {
	dao.IdTitle
	Bounds  geo.Bbox   `json:"bounds"`
	Region  dao.Region `json:"region"`
	Visible bool       `json:"visible"`
}

type RiverPageDto struct {
	dao.IdTitle
	Region        dao.Region             `json:"region"`
	Description   string                 `json:"description"`
	Reports       []VoyageReportListDto  `json:"reports"`
	Transfers     []dao.Transfer         `json:"transfers"`
	HasCamps      bool                   `json:"has_camps"`
	Imgs          []ImgWithSpot          `json:"imgs"`
	Videos        []ImgWithSpot          `json:"videos"`
	PdfUrl        string                 `json:"pdf"`
	HtmlUrl       string                 `json:"html"`
	Props         map[string]interface{} `json:"props"`
	MaxCategory   model.SportCategory    `json:"max_category"`
	AvgCategory   model.SportCategory    `json:"avg_category"` // min category of 3 hardest spots
	HasImpassible bool                   `json:"has_impassible"`
	WeatherPoint  *geo.Point             `json:"weather_point"`
	SearchQuery   string                 `json:"search_query"`
	ResourceBase  string                 `json:"resource_base"`
}

type ImgWithSpot struct {
	dao.Img
	SpotTitle string `json:"spot_title"`
}

func (this *RiverHandler) GetRiverCard(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	river, err := this.TileDao.GetRiver(riverId, 1)
	if err != nil {
		OnError500(w, err, "Can not select river")
		return
	}

	reports, err := this.VoyageReportDao.ByRiver(river.Id, MAX_REPORTS_PER_SOURCE)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not select reports for river %d", river.Id))
		return
	}
	reportsList := this.groupReports(reports, river)

	imgs := []ImgWithSpot{}
	videos := []ImgWithSpot{}
	for i := 0; i < len(river.Spots); i++ {
		wasVideo := false
		wasImg := false
		for j := 0; j < len(river.Spots[i].Images); j++ {
			if !wasVideo && river.Spots[i].Images[j].Type == dao.IMAGE_TYPE_VIDEO {
				videos = append(videos, ImgWithSpot{Img: river.Spots[i].Images[j], SpotTitle: river.Spots[i].Title})
				wasVideo = true
			}
			if !wasImg && river.Spots[i].Images[j].Type == dao.IMAGE_TYPE_IMAGE {
				img := river.Spots[i].Images[j]
				this.processForWeb(&img)
				imgs = append(imgs, ImgWithSpot{Img: img, SpotTitle: river.Spots[i].Title})
				wasImg = true
			}
			if wasImg && wasVideo {
				break
			}
		}
	}

	var weatherPoint *geo.Point = nil
	meteoPointId, meteoPointFound := river.Props["meteo_point"]
	if meteoPointFound && meteoPointId != nil {
		p, err := this.MeteoPointDao.Find(int64(meteoPointId.(float64)))
		if err != nil {
			log.Errorf("Failed to find meteo point for river id=%d: %v", river.Id, err)
		} else {
			weatherPoint = &(p.Point)
		}
	}

	transfers, err := this.TransferDao.ByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not list transfers for river %d", riverId))
		return
	}

	hasCamps, err := this.CampDao.ExistsByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not detect river %d has camps", riverId))
		return
	}

	riverCats := dao.CalculateClusterCategory(river.Spots)

	dto := RiverPageDto{
		IdTitle:       river.IdTitle,
		Region:        river.Region,
		Description:   river.Description,
		Props:         river.Props,
		Reports:       reportsList,
		Transfers:     transfers,
		HasCamps:      hasCamps,
		Imgs:          imgs,
		Videos:        videos,
		PdfUrl:        this.getRiverPassportUrl(
			&river,
			this.RiverPassportPdfStorageParams,
			this.RiverPassportPdfStorage,
			),
		HtmlUrl:       this.getRiverPassportUrl(
			&river,
			this.RiverPassportHtmlStorageParams,
			this.RiverPassportHtmlStorage,
			),
		MaxCategory:   model.SportCategory{Category: riverCats.Max},
		AvgCategory:   model.SportCategory{Category: riverCats.Avg},
		HasImpassible: riverCats.HasImpassible,
		WeatherPoint:  weatherPoint,
		SearchQuery: strings.Join([]string{
			url.QueryEscape("река"),
			url.QueryEscape(river.Title),
			url.QueryEscape("сплав")}, "+"),
		ResourceBase: this.ResourceBase,
	}
	w.Write([]byte(JsonStr(dto, "{}")))
}

func (this *RiverHandler) groupReports(reports []dao.VoyageReport, river dao.RiverWithSpotsExt) []VoyageReportListDto {
	reportDtos := make(map[string][]VoyageReportDto)
	for _, report := range reports {
		reportDtos[report.Source] = append(reportDtos[report.Source], VoyageReportDto{
			Id:            report.Id,
			Url:           report.Url,
			Title:         report.Title,
			Author:        report.Author,
			Year:          report.GetDateOfTripYear(0),
			SourceLogoUrl: this.ResourceBase + "/img/report_sources/" + strings.ToLower(report.Source) + ".png",
		})
	}
	reportsListBuilder := ReportsListBuilder{
		source:    reportDtos,
		processed: make(map[string]bool),

		// Serialize empty report list as empty array not null
		// https://apoorvam.github.io/blog/2017/golang-json-marshal-slice-as-empty-array-not-null/
		reportsList: make([]VoyageReportListDto, 0),
	}
	for _, s := range REPORT_SOURCES {
		reportsListBuilder.addReportDtos(s.Id, s.Title)
	}
	reportsListBuilder.others()
	return reportsListBuilder.reportsList
}

type ReportsListBuilder struct {
	reportsList []VoyageReportListDto
	source      map[string][]VoyageReportDto
	processed   map[string]bool
}

func (this *ReportsListBuilder) addReportDtos(source string, alias string) {
	reports, found := this.source[source]
	if found {
		this.reportsList = append(this.reportsList, VoyageReportListDto{
			Source:  alias,
			Reports: reports,
		})
		this.processed[source] = true
	}
}
func (this *ReportsListBuilder) others() {
	for s, l := range this.source {
		_, processed := this.processed[s]
		if !processed {
			this.reportsList = append(this.reportsList, VoyageReportListDto{
				Source:  s,
				Reports: l,
			})
		}
	}
}

func (this *RiverHandler) GetVisibleRiversLite(w http.ResponseWriter, req *http.Request) {
	bbox, err := BboxFormValue(w, req)
	if err != nil {
		return
	}

	catFilter := 0
	catFilterStr := req.FormValue("max_cat")
	if catFilterStr != "" {
		catFilter, err = strconv.Atoi(catFilterStr)
		if err != nil {
			log.Warnf("Incorrect max_cat parameter %s", catFilterStr)
			catFilter = 0
		}
	}

	featureToggles := toggles.ParseFeatureTogglesOrFallback(req, this.UserDao)
	showUnpublished, ctx := featureToggles.GetShowUnpublished(req.Context())
	if req.Context() != ctx {
		req = req.WithContext(ctx)
	}

	rivers, err := this.TileDao.ListRiversWithBounds(bbox, RIVER_LIST_LIMIT, showUnpublished)
	if err != nil {
		OnError500(w, err, "Can not select rivers")
		return
	}

	riversWithReports := make([]RiverListLiteDto, 0, len(rivers))
	for i := 0; i < len(rivers); i++ {
		river := &rivers[i]
		if len(river.Spots) == 0 {
			continue
		}
		maxCat := 0
		bounds := geo.Bbox{
			X1: 360.0,
			Y1: 360.0,
			X2: -360.0,
			Y2: -360.0,
		}
		for j := 0; j < len(river.Spots); j++ {
			cat := river.Spots[j].Category.Category
			if maxCat < cat {
				maxCat = cat
			}
			bounds.AddPointOrLine(river.Spots[j].Point)
		}
		if maxCat < catFilter {
			continue
		}

		riversWithReports = append(riversWithReports, RiverListLiteDto{
			IdTitle: dao.IdTitle{Id: river.Id, Title: river.Title},
			Region:  dao.Region{Id: river.RegionId, CountryId: river.CountryId},
			Bounds:  bounds.WithMargins(RIVER_BOUNDS_MARGINS_RATIO),
			Visible: river.Visible,
		})
	}
	w.Write([]byte(JsonStr(riversWithReports, "[]")))
}

func (this *RiverHandler) GetVisibleRivers(w http.ResponseWriter, req *http.Request) {
	bbox, err := BboxFormValue(w, req)
	if err != nil {
		return
	}

	rivers, err := this.RiverDao.ListRiversWithBounds(bbox, RIVER_LIST_LIMIT, false)
	if err != nil {
		OnError500(w, err, "Can not select rivers")
		return
	}

	riversWithReports := make([]RiverListDto, len(rivers))
	for i := 0; i < len(rivers); i++ {
		river := &rivers[i]
		river.Bounds = river.Bounds.WithMargins(RIVER_BOUNDS_MARGINS_RATIO)

		reports, err := this.VoyageReportDao.ByRiver(river.Id, MAX_REPORTS_PER_SOURCE)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not select reports for river %d", river.Id))
			return
		}
		reportDtos := make([]VoyageReportDto, len(reports))
		for j, report := range reports {
			reportDtos[j] = VoyageReportDto{
				Id:            report.Id,
				Url:           report.Url,
				Title:         report.Title,
				Author:        report.Author,
				Year:          report.GetDateOfTripYear(0),
				SourceLogoUrl: this.ResourceBase + "/img/report_sources/" + strings.ToLower(report.Source) + ".png",
			}
		}

		riversWithReports[i] = RiverListDto{
			RiverTitle: *river,
			Reports:    reportDtos,
			PdfUrl:     this.getRiverPassportUrl(river, this.RiverPassportPdfStorageParams, nil),
			HtmlUrl:    this.getRiverPassportUrl(river, this.RiverPassportHtmlStorageParams, nil),
		}

	}
	w.Write([]byte(JsonStr(riversWithReports, "[]")))
}

func (this *RiverHandler) getRiverPassportUrl(
	river HasPropertiesAndId,
	storageParams config.BlobStorageParams,
	storage blob.BlobStorage,
	) string {
	export, found := river.GetProperties()["export_pdf"]
	if found && export.(bool) {
		if storage == nil {
			return createRiverPassportUrl(river, storageParams.UrlBase)
		}

		key := fmt.Sprintf("%d.%s", river.GetId(), storageParams.Suffix)
		if exists, err := storage.Exists(key); err == nil && exists {
			return createRiverPassportUrl(river, storageParams.UrlBase)
		}
	}
	return ""
}

func createRiverPassportUrl(river HasPropertiesAndId, base string) string {
	return fmt.Sprintf(base, river.GetId())
}

type HasPropertiesAndId interface {
	GetId() int64
	GetProperties() map[string]interface{}
}
