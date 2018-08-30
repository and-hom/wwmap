package main

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
	"encoding/json"
)

type Handler struct {
	App
}

func (this *Handler) TileHandler(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, GET, OPTIONS)

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

func (this *Handler) RefSites(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, "GET, OPTIONS")
	JsonResponse(w)

	refs := this.refererStorage.List()
	bytes, err := json.Marshal(refs)
	if err != nil {
		OnError500(w, err, "Can not marshal json")
		return
	}
	w.Write(bytes)
}


