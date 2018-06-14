package main

import (
	"net/http"
	"fmt"
	"strconv"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/dao"
	"strings"
)

type RiverHandler struct {
	Handler
	resourceBase string
}

const MAX_REPORTS_PER_SOURCE = 5

func (this *RiverHandler) GetNearestRivers(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, "GET")
	lat_s := r.FormValue("lat")
	lat, err := strconv.ParseFloat(lat_s, 64)
	if err != nil {
		OnError(w, err, fmt.Sprintf("Can not parse lat parameter: %s", lat_s), 400)
		return
	}
	lon_s := r.FormValue("lon")
	lon, err := strconv.ParseFloat(lon_s, 64)
	if err != nil {
		OnError(w, err, fmt.Sprintf("Can not parse lon parameter: %s", lon_s), 400)
		return
	}
	point := Point{Lat:lat, Lon:lon}
	rivers, err := this.riverDao.NearestRivers(point, 5)
	if err != nil {
		OnError500(w, err, "Can not select rivers")
		return
	}
	w.Write([]byte(this.JsonStr(rivers, "[]")))
}

type VoyageReportDto struct {
	Id            int64 `json:"id"`
	Title         string `json:"title"`
	Year          int `json:"year"`
	Url           string `json:"url"`
	SourceLogoUrl string `json:"source_logo_url"`
}

type RiverWithReports struct {
	dao.RiverTitle
	Reports []VoyageReportDto `json:"reports"`
}

func (this *RiverHandler) GetVisibleRivers(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, "GET")

	bbox, err := this.bboxFormValue(w, req)
	if err != nil {
		return
	}

	rivers, err := this.riverDao.ListRiversWithBounds(bbox, 30)
	if err != nil {
		OnError500(w, err, "Can not select rivers")
		return
	}

	riversWithReports := make([]RiverWithReports, len(rivers))
	for i := 0; i < len(rivers); i++ {
		river := &rivers[i]
		river.Bounds = river.Bounds.WithMargins(0.05)

		reports, err := this.voyageReportDao.List(river.Id, MAX_REPORTS_PER_SOURCE)
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
				Year:report.DateOfTrip.Year(),
				SourceLogoUrl:this.resourceBase + "/img/report_sources/" + strings.ToLower(report.Source) + ".png",
			}
		}

		riversWithReports[i] = RiverWithReports{*river, reportDtos}

	}
	w.Write([]byte(this.JsonStr(riversWithReports, "[]")))
}
