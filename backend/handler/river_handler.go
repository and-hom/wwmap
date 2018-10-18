package handler

import (
	"net/http"
	"fmt"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/lib/dao"
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
}

const MAX_REPORTS_PER_SOURCE = 5

type VoyageReportDto struct {
	Id            int64 `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Year          int `json:"year"`
	Url           string `json:"url"`
	SourceLogoUrl string `json:"source_logo_url"`
}

type RiverListDto struct {
	dao.RiverTitle
	Reports []VoyageReportDto `json:"reports"`
	PdfUrl  string `json:"pdf"`
	HtmlUrl string `json:"html"`
}

func (this *RiverHandler) GetVisibleRivers(w http.ResponseWriter, req *http.Request) {
	bbox, err := this.bboxFormValue(w, req)
	if err != nil {
		return
	}

	rivers, err := this.RiverDao.ListRiversWithBounds(bbox, 30, false)
	if err != nil {
		OnError500(w, err, "Can not select rivers")
		return
	}

	riversWithReports := make([]RiverListDto, len(rivers))
	for i := 0; i < len(rivers); i++ {
		river := &rivers[i]
		river.Bounds = river.Bounds.WithMargins(0.05)

		reports, err := this.VoyageReportDao.List(river.Id, MAX_REPORTS_PER_SOURCE)
		if err != nil {
			OnError500(w, err, fmt.Sprintf("Can not select reports for river %d", river.Id))
			return
		}
		reportDtos := make([]VoyageReportDto, len(reports))
		for j, report := range reports {
			reportDtos[j] = VoyageReportDto{
				Id:report.Id,
				Url:report.Url,
				Title:report.Title,
				Author:report.Author,
				Year:report.DateOfTrip.Year(),
				SourceLogoUrl:this.ResourceBase + "/img/report_sources/" + strings.ToLower(report.Source) + ".png",
			}
		}

		riversWithReports[i] = RiverListDto{
			RiverTitle: *river,
			Reports: reportDtos,
			PdfUrl: this.getRiverPassportUrl(river, this.RiverPassportPdfUrlBase),
			HtmlUrl:this.getRiverPassportUrl(river, this.RiverPassportHtmlUrlBase),
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
