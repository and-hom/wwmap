package main

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
)

type Handler struct {
	App
}

func (this *Handler) TileHandler(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, "GET, OPTIONS")

	callback, bbox, err := this.tileParams(w, req)
	if err != nil {
		return
	}
	tracks := this.storage.ListTracks(bbox)
	points := this.storage.ListPoints(bbox)

	featureCollection := MkFeatureCollection(append(pointsToYmaps(points), tracksToYmaps(tracks)...))
	log.Infof("Found %d", len(featureCollection.Features))

	w.Write(this.JsonpAnswer(callback, featureCollection, "{}"))
}


