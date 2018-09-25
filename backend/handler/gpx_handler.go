package handler

import (
	"github.com/gorilla/mux"
	"strconv"
	"net/http"
	"fmt"
	"github.com/ptrv/go-gpx"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/lib/util"
)

type GpxHandler struct{ App };

func (this *GpxHandler) Init(r *mux.Router) {
	this.Register(r, "/gpx/river/{id}", HandlerFunctions{Get: this.DownloadGpx})
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
		OnError500(w, err, fmt.Sprintf("Can not read whitewater points for river %s", id))
		return
	}

	waypoints := make([]gpx.Wpt, len(whitewaterPoints))
	for i := 0; i < len(whitewaterPoints); i++ {
		whitewaterPoint := whitewaterPoints[i]
		waypoints[i] = gpx.Wpt{
			Lat: whitewaterPoint.Point.Lat,
			Lon: whitewaterPoint.Point.Lon,
			Cmt: whitewaterPoint.Comment,
		}
		categoryString := util.HumanReadableCategoryNameWithBrackets(whitewaterPoint.Category, transliterate)

		titleString := whitewaterPoint.Title
		if transliterate {
			titleString = util.CyrillicToTranslit(whitewaterPoint.Title)
		}

		waypoints[i].Name = categoryString + titleString
	}
	if len(whitewaterPoints) == 0 {
		OnError(w, nil, fmt.Sprintf("No whitewater points found for river with id %d", id), http.StatusNotFound)
		return
	}
	gpxData := gpx.Gpx{
		Waypoints: waypoints,
		Creator: "wwmap",
	}
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.gpx\"", whitewaterPoints[0].RiverTitle))
	w.Header().Add("Content-Type", "application/gpx+xml")

	xmlBytes := gpxData.ToXML()
	w.Write(xmlBytes)
}
