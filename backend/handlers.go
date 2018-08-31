package main

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/mux"
)

type ApiHandler interface {
	Init(*mux.Router)
}

type Handler struct {
	App
}

type HandlerFunction func(http.ResponseWriter, *http.Request)
type HandlerFunctions struct {
	get    HandlerFunction
	post   HandlerFunction
	put    HandlerFunction
	delete HandlerFunction
}

func (this *HandlerFunctions) CorsMethods() []string {
	corsMethods := []string{}
	if this.get != nil {
		corsMethods = append(corsMethods, GET)
	}
	if this.post != nil {
		corsMethods = append(corsMethods, POST)
	}
	if this.put != nil {
		corsMethods = append(corsMethods, PUT)
	}
	if this.delete != nil {
		corsMethods = append(corsMethods, DELETE)
	}
	return corsMethods
}

func (this *Handler) Register(r *mux.Router, path string, handlerFunctions HandlerFunctions) {
	corsMethods := handlerFunctions.CorsMethods()

	r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		this.CorOptionsStub(w, r, corsMethods)
	}).Methods(OPTIONS)

	this.registerOne(r, path, GET, handlerFunctions.get, corsMethods)
	this.registerOne(r, path, PUT, handlerFunctions.put, corsMethods)
	this.registerOne(r, path, POST, handlerFunctions.post, corsMethods)
	this.registerOne(r, path, DELETE, handlerFunctions.delete, corsMethods)
}

func (this *Handler) registerOne(r *mux.Router, path string, method string, handlerFunction HandlerFunction, corsMethods []string) {
	if handlerFunction == nil {
		return
	}
	r.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		CorsHeaders(w, corsMethods...)
		handlerFunction(w, r)
	}).Methods(method)
}

func (this *Handler) TileHandler(w http.ResponseWriter, req *http.Request) {
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

