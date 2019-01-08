package handler

import (
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

type RiverHandler struct {
	App
	ResourceBase             string
	RiverPassportPdfUrlBase  string
	RiverPassportHtmlUrlBase string
}

func (this *RiverHandler) Init() {
	this.Register("/visible-rivers", HandlerFunctions{Get: this.GetVisibleRivers})
	this.Register("/visible-rivers-light", HandlerFunctions{Get: this.GetVisibleRiversLight})
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

type RiverListDto struct {
	dao.RiverTitle
	Reports []VoyageReportDto `json:"reports"`
	PdfUrl  string            `json:"pdf"`
	HtmlUrl string            `json:"html"`
}

type RiverPageDto struct {
	dao.RiverTitle
	Description string                       `json:"description"`
	Reports     map[string][]VoyageReportDto `json:"reports"`
	Imgs        []dao.Img                    `json:"imgs"`
	PdfUrl      string                       `json:"pdf"`
	HtmlUrl     string                       `json:"html"`
}

func (this *RiverHandler) GetRiverCard(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)
	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	river, err := this.RiverDao.Find(riverId)
	if err != nil {
		OnError500(w, err, "Can not select river")
		return
	}

	reports, err := this.VoyageReportDao.List(river.Id, MAX_REPORTS_PER_SOURCE)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not select reports for river %d", river.Id))
		return
	}
	reportDtos := make(map[string][]VoyageReportDto)
	for _, report := range reports {
		reportDtos[report.Source] = append(reportDtos[report.Source], VoyageReportDto{
			Id:            report.Id,
			Url:           report.Url,
			Title:         report.Title,
			Author:        report.Author,
			Year:          report.DateOfTrip.Year(),
			SourceLogoUrl: this.ResourceBase + "/img/report_sources/" + strings.ToLower(report.Source) + ".png",
		})
	}

	imgs, err := this.ImgDao.ListMainByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not select images for river %d", river.Id))
		return
	}
	for i := 0; i < len(imgs); i++ {
		this.processForWeb(&(imgs[i]))
	}

	dto := RiverPageDto{
		RiverTitle:  river.RiverTitle,
		Description: river.Description,
		Reports:     reportDtos,
		Imgs:        imgs,
		PdfUrl:      this.getRiverPassportUrl(&river.RiverTitle, this.RiverPassportPdfUrlBase),
		HtmlUrl:     this.getRiverPassportUrl(&river.RiverTitle, this.RiverPassportHtmlUrlBase),
	}
	w.Write([]byte(this.JsonStr(dto, "{}")))
}

func (this *RiverHandler) GetVisibleRiversLight(w http.ResponseWriter, req *http.Request) {
	bbox, err := this.bboxFormValue(w, req)
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
		river.Props = nil

		riversWithReports[i] = RiverListDto{
			RiverTitle: *river,
		}

	}
	w.Write([]byte(this.JsonStr(riversWithReports, "[]")))
}

func (this *RiverHandler) GetVisibleRivers(w http.ResponseWriter, req *http.Request) {
	bbox, err := this.bboxFormValue(w, req)
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

		reports, err := this.VoyageReportDao.List(river.Id, MAX_REPORTS_PER_SOURCE)
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
				Year:          report.DateOfTrip.Year(),
				SourceLogoUrl: this.ResourceBase + "/img/report_sources/" + strings.ToLower(report.Source) + ".png",
			}
		}

		riversWithReports[i] = RiverListDto{
			RiverTitle: *river,
			Reports:    reportDtos,
			PdfUrl:     this.getRiverPassportUrl(river, this.RiverPassportPdfUrlBase),
			HtmlUrl:    this.getRiverPassportUrl(river, this.RiverPassportHtmlUrlBase),
		}

	}
	w.Write([]byte(this.JsonStr(riversWithReports, "[]")))
}

func (this *RiverHandler) getRiverPassportUrl(river *dao.RiverTitle, base string) string {
	export, found := river.Props["export_pdf"]
	if found && export.(bool) {
		return fmt.Sprintf(base, river.Id)
	}
	return ""
}
