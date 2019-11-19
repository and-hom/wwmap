package handler

import (
	"encoding/csv"
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/model"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/gorilla/mux"
	"github.com/ptrv/go-gpx"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"net/http"
	"strconv"
)

type DownloadsHandler struct{ App }

func (this *DownloadsHandler) Init() {
	this.Register("/downloads/river/{id}/gpx", HandlerFunctions{Get: this.DownloadGpx})
	this.Register("/downloads/river/{id}/csv", HandlerFunctions{Get: this.DownloadCsv})
}

func (this *DownloadsHandler) DownloadGpx(w http.ResponseWriter, req *http.Request) {
	this.Download(w, req, this.toGpxBytes)
}

func (this *DownloadsHandler) DownloadCsv(w http.ResponseWriter, req *http.Request) {
	this.Download(w, req, this.toCsvBytes)
}

func (this *DownloadsHandler) Download(w http.ResponseWriter, req *http.Request, responseBodyWriter func(http.ResponseWriter, string, []dao.WhiteWaterPointWithRiverTitle, bool) error) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	transliterate := req.FormValue("tr") != ""

	whitewaterPoints, err := this.WhiteWaterDao.ListByRiver(id)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not read whitewater points for river %d", id))
		return
	}
	if len(whitewaterPoints) == 0 {
		OnError(w, nil, fmt.Sprintf("No whitewater points found for river with id %d", id), http.StatusNotFound)
		return
	}

	filename := whitewaterPoints[0].RiverTitle
	if transliterate {
		filename = util.CyrillicToTranslit(filename)
	}

	responseBodyWriter(w, filename, whitewaterPoints, transliterate)
}

func (this *DownloadsHandler) toGpxBytes(w http.ResponseWriter, filename string,
	whitewaterPoints []dao.WhiteWaterPointWithRiverTitle, transliterate bool) error {

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.gpx\"", filename))
	w.Header().Add("Content-Type", "application/gpx+xml")

	waypoints := make([]gpx.Wpt, 0, len(whitewaterPoints))
	for i := 0; i < len(whitewaterPoints); i++ {
		whitewaterPoint := whitewaterPoints[i]

		categoryString := util.HumanReadableCategoryNameWithBrackets(whitewaterPoint.Category, transliterate)

		titleString := whitewaterPoint.Title
		if transliterate {
			titleString = util.CyrillicToTranslit(whitewaterPoint.Title)
		}

		if whitewaterPoint.Point.Point != nil {
			waypoints = append(waypoints, gpx.Wpt{
				Lat:  whitewaterPoint.Point.Point.Lat,
				Lon:  whitewaterPoint.Point.Point.Lon,
				Cmt:  whitewaterPoint.Comment,
				Name: categoryString + titleString,
			})
		} else {
			pBegin := (*whitewaterPoint.Point.Line)[0]

			endIdx := len(*whitewaterPoint.Point.Line) - 1
			pEnd := (*whitewaterPoint.Point.Line)[endIdx]

			waypoints = append(waypoints, gpx.Wpt{
				Lat:  pBegin.Lat,
				Lon:  pBegin.Lon,
				Cmt:  whitewaterPoint.Comment,
				Name: categoryString + titleString + orderString(0, whitewaterPoint, transliterate),
			})
			waypoints = append(waypoints, gpx.Wpt{
				Lat:  pEnd.Lat,
				Lon:  pEnd.Lon,
				Cmt:  whitewaterPoint.Comment,
				Name: categoryString + titleString + orderString(endIdx, whitewaterPoint, transliterate),
			})
		}
	}
	gpxData := gpx.Gpx{
		Waypoints: waypoints,
		Creator:   "wwmap",
	}

	_, err := w.Write(gpxData.ToXML())
	return err
}

var header_ru = []string{"#", "Название", "Категория", "Широта", "Долгота"}
var header_en = []string{"#", "Name", "Category", "Lat", "Lon"}

func (this *DownloadsHandler) toCsvBytes(w http.ResponseWriter, filename string,
	whitewaterPoints []dao.WhiteWaterPointWithRiverTitle, transliterate bool) error {

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", filename))
	w.Header().Add("Content-Type", "application/vnd.ms-excel; charset=windows-1251")

	encodedWriter := transform.NewWriter(w, charmap.Windows1251.NewEncoder())

	writer := csv.NewWriter(encodedWriter)
	defer writer.Flush()
	writer.Comma = ';'

	if transliterate {
		writer.Write(header_en)
	} else {
		writer.Write(header_ru)
	}

	idx := 0
	for _, spot := range whitewaterPoints {
		title := spot.Title
		if transliterate {
			title = util.CyrillicToTranslit(title)
		}

		if spot.Point.Point != nil {
			if err := this.writeCsvLine(writer, idx, title, spot.Category, spot.Point.Point, transliterate); err != nil {
				return err
			}
			idx++
		} else {
			pBegin := (*spot.Point.Line)[0]
			endIdx := len(*spot.Point.Line) - 1
			pEnd := (*spot.Point.Line)[endIdx]

			if err := this.writeCsvLine(writer, idx, title+orderString(0, spot, transliterate), spot.Category, &pBegin, transliterate); err != nil {
				return err
			}
			idx++

			if err := this.writeCsvLine(writer, idx, title+orderString(endIdx, spot, transliterate), spot.Category, &pEnd, transliterate); err != nil {
				return err
			}
			idx++
		}
	}
	return nil
}

func (this *DownloadsHandler) writeCsvLine(writer *csv.Writer, idx int, title string, category model.SportCategory, p *geo.Point, transliterate bool) error {
	return writer.Write([]string{
		strconv.Itoa(idx + 1),
		title,
		util.HumanReadableCategoryName(category, transliterate),
		strconv.FormatFloat(p.Lat, 'f', 12, 64),
		strconv.FormatFloat(p.Lon, 'f', 12, 64),
	})
}

func orderString(i int, spot dao.WhiteWaterPointWithRiverTitle, transliterate bool) string {
	if transliterate {
		return orderStringEn(i, spot)
	} else {
		return orderStringRu(i, spot)
	}
}

func orderStringRu(i int, spot dao.WhiteWaterPointWithRiverTitle) string {
	switch i {
	case 0:
		return " (Нач.)"
	case len(*spot.Point.Line) - 1:
		return " (Кон.)"
	default:
		return fmt.Sprintf(" (тчк. %d)", i)
	}
}

func orderStringEn(i int, spot dao.WhiteWaterPointWithRiverTitle) string {
	switch i {
	case 0:
		return " (Na4.)"
	case len(*spot.Point.Line) - 1:
		return " (Kon.)"
	default:
		return fmt.Sprintf(" (t4k. %d)", i)
	}
}
