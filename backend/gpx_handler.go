package main

import (
	"github.com/gorilla/mux"
	"strconv"
	"net/http"
	"fmt"
	"github.com/ptrv/go-gpx"
)

type GpxHandler struct{ Handler };

func (this *GpxHandler) DownloadGpx(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET")
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		this.onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	whitewaterPoints, err := this.whiteWaterDao.ListWhiteWaterPointsByRiver(id)
	if err != nil {
		this.onError500(w, err, fmt.Sprintf("Can not read whitewater points for river %s", id))
		return
	}

	waypoints := make([]gpx.Wpt, len(whitewaterPoints))
	for i := 0; i < len(whitewaterPoints); i++ {
		whitewaterPoint := whitewaterPoints[i]
		waypoints[i] = gpx.Wpt{
			Lat: whitewaterPoint.Point.Lat,
			Lon: whitewaterPoint.Point.Lon,
			Name: whitewaterPoint.Title,
			Cmt: whitewaterPoint.Comment,
		}
	}
	if len(whitewaterPoints) == 0 {
		this.onError(w, nil, fmt.Sprintf("No whitewater points found for river with id %d", id), http.StatusNotFound)
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
