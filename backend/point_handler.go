package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
)

type PointHandler struct {
	Handler
}

func (this *PointHandler) GetPoint(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, POST, GET, OPTIONS, PUT, DELETE)

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	this.writePointToResponse(id, w)
}

func (this *PointHandler) writePointToResponse(id int64, w http.ResponseWriter) {
	eventPoint := EventPoint{}
	found, err := this.storage.FindEventPoint(id, &eventPoint)
	if err != nil {
		OnError500(w, err, "Can not find")
		return
	}
	if !found {
		OnError(w, fmt.Errorf("Point with id %d does not exist", id), "Not found", http.StatusNotFound)
	}
	bytes, err := json.Marshal(eventPoint)
	if err != nil {
		OnError500(w, err, "Can not marshal")
		return
	}
	w.Write(bytes)
}

func (this *PointHandler) DelPoint(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, POST, GET, OPTIONS, PUT, DELETE)
	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	err = this.storage.DeleteEventPoint(id)
	if err != nil {
		OnError500(w, err, "Can not delete")
		return
	}
}

func (this *PointHandler) EditPoint(w http.ResponseWriter, r *http.Request) {
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

	eventPoint, _, err := this.parsePointForm(w, r)
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}
	eventPoint.Id = id

	err = this.storage.UpdateEventPoint(eventPoint)
	if err != nil {
		OnError500(w, err, "Can not edit point	")
		return
	}
	this.writePointToResponse(id, w)
}

func (this *PointHandler) AddPoint(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, POST, GET, OPTIONS, PUT, DELETE)
	err := r.ParseForm()
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}

	eventPoint, routeId, err := this.parsePointForm(w, r)
	if err != nil {
		return
	}

	id, err := this.storage.AddEventPoint(routeId, eventPoint)

	if err != nil {
		OnError500(w, err, "Can not insert")
		return
	}

	w.Write([]byte(strconv.FormatInt(id, 10)))
}

func (this *PointHandler) parsePointForm(w http.ResponseWriter, r *http.Request) (EventPoint, int64, error) {
	route_id, err := strconv.ParseInt(r.FormValue("route_id"), 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return EventPoint{}, 0, err
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	pType, err := ParseEventPointType(r.FormValue("type"))
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return EventPoint{}, 0, err
	}
	point := Point{};
	err = json.Unmarshal([]byte(r.FormValue("point")), &point)
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return EventPoint{}, 0, err
	}

	eventPoint := EventPoint{
		Type:pType,
		Title:title,
		Content:content,
		Point:point,
		Time:JSONTime(time.Now()),
	}

	return eventPoint, route_id, nil
}

