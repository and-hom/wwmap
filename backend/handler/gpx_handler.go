package handler

import (
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/gorilla/mux"
	"github.com/ptrv/go-gpx"
	"net/http"
	"strconv"
)

type GpxHandler struct{ App }

func (this *GpxHandler) Init() {
	this.Register("/gpx/river/{id}", HandlerFunctions{Get: this.DownloadGpx})
}

func (this *GpxHandler) DownloadGpx(w http.ResponseWriter, req *http.Request) {
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
	if len(whitewaterPoints) == 0 {
		OnError(w, nil, fmt.Sprintf("No whitewater points found for river with id %d", id), http.StatusNotFound)
		return
	}
	gpxData := gpx.Gpx{
		Waypoints: waypoints,
		Creator:   "wwmap",
	}

	filename := whitewaterPoints[0].RiverTitle
	if transliterate {
		filename = util.CyrillicToTranslit(filename)
	}

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.gpx\"", filename))
	w.Header().Add("Content-Type", "application/gpx+xml")

	xmlBytes := gpxData.ToXML()
	w.Write(xmlBytes)
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
