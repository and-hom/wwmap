package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"fmt"
	"strconv"
	. "github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/model"
	. "github.com/and-hom/wwmap/lib/http"
)

type RouteHandler struct {
	Handler
}

func (this *RouteHandler) EditRoute(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, POST, GET, OPTIONS, PUT, DELETE)
	err := r.ParseForm()
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	route, err := this.parseRouteForm(w, r)
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}
	route.Id = id

	err = this.storage.UpdateRoute(route)
	if err != nil {
		OnError500(w, err, "Can not edit route")
		return
	}
	this.writeRouteToResponse(id, w)
}

func (this *RouteHandler) DelRoute(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, POST, GET, OPTIONS, PUT, DELETE)

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = this.storage.DeleteTracksForRoute(id)
	if err != nil {
		OnError500(w, err, "Can not remove tracks")
		return
	}
	err = this.storage.DeleteEventPointsForRoute(id)
	if err != nil {
		OnError500(w, err, "Can not remove points")
		return
	}
	err = this.storage.DeleteRoute(id)
	if err != nil {
		OnError500(w, err, "Can not remove route")
		return
	}
}

func (this *RouteHandler) writeRouteToResponse(id int64, w http.ResponseWriter) {
	route := Route{}
	found, err := this.storage.FindRoute(id, &route)
	if err != nil {
		OnError500(w, err, "Can not find")
		return
	}
	if !found {
		OnError(w, fmt.Errorf("Route with id %d does not exist", id), "Not found", http.StatusNotFound)
	}
	bytes, err := json.Marshal(route)
	if err != nil {
		OnError500(w, err, "Can not marshal")
		return
	}
	w.Write(bytes)
}

func (this *RouteHandler) RouteEditorPageHandler(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, GET, OPTIONS)

	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	route := Route{}
	found, err := this.storage.FindRoute(id, &route)
	if found {
		tracks := this.storage.FindTracksForRoute(route.Id)
		points := this.storage.FindEventPointsForRoute(route.Id)
		bytes, err := json.Marshal(RouteEditorPage{
			Title:route.Title,
			Bounds: Bounds(tracks, points),
			Tracks: tracks,
			EventPoints:points,
			Category:route.Category,
		})
		if err != nil {
			OnError500(w, err, "Can not serialize track")
			return
		}
		w.Write(bytes)
	} else {
		OnError500(w, err, "Track not found")
	}
}

func (this *RouteHandler) GetVisibleRoutes(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, GET)

	bbox, err := this.bboxFormValue(w, req)
	if err != nil {
		return
	}

	routes := this.storage.ListRoutes(bbox)
	log.Infof("Found %d", len(routes))

	routeInfos := make([]RouteEditorPage, len(routes))
	for i := 0; i < len(routes); i++ {
		route := routes[i]
		routeInfos[i] = RouteEditorPage{
			Id:route.Id,
			Title:route.Title,
			Tracks:this.storage.FindTracksForRoute(route.Id),
			EventPoints:this.storage.FindEventPointsForRoute(route.Id),
			Category: route.Category,
		}
	}

	w.Write([]byte(this.JsonStr(routeInfos, "[]")))
}

func (this *RouteHandler) SingleRouteTileHandler(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, GET, OPTIONS)

	callback := req.FormValue("callback")
	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	tracks := this.storage.FindTracksForRoute(id)
	points := this.storage.FindEventPointsForRoute(id)

	featureCollection := geo.MkFeatureCollection(append(pointsToYmaps(points), tracksToYmaps(tracks)...))
	log.Infof("Found %d", len(featureCollection.Features))

	w.Write(this.JsonpAnswer(callback, featureCollection, "{}"))
}

func (this *RouteHandler) parseRouteForm(w http.ResponseWriter, r *http.Request) (Route, error) {
	title := r.FormValue("title")
	category := model.SportCategory{}
	err := json.Unmarshal([]byte(r.FormValue("category")), &category)
	if err != nil {
		return Route{}, err
	}
	route := Route{
		Title:title,
		Category: category,
	}
	return route, nil
}


