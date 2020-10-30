package handler

import (
	"encoding/json"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"net/http"
)

type MeteoHandler struct {
	App
}

func (this *MeteoHandler) Init() {
	this.Register("/meteo-point", HandlerFunctions{
		Get:  this.ForRoles(this.ListMeteoPoints, dao.ADMIN, dao.EDITOR),
		Put:  this.ForRoles(this.AddMeteoPoint, dao.ADMIN, dao.EDITOR),
		Post: this.ForRoles(this.AddMeteoPoint, dao.ADMIN, dao.EDITOR),
	})
}

func (this *MeteoHandler) AddMeteoPoint(w http.ResponseWriter, req *http.Request) {
	point := dao.MeteoPoint{}
	err := json.NewDecoder(req.Body).Decode(&point)
	if err != nil {
		OnError(w, err, "Can't parse request body", http.StatusBadRequest)
		return
	}
	p, err := this.MeteoPointDao.Insert(point)
	if err != nil {
		OnError500(w, err, "Can't insert meteo point")
		return
	}
	JsonAnswer(w, p)
}

func (this *MeteoHandler) ListMeteoPoints(w http.ResponseWriter, req *http.Request) {
	this.listMeteoPoints(w)
}

func (this *MeteoHandler) listMeteoPoints(w http.ResponseWriter) {
	points, err := this.MeteoPointDao.List()
	if err != nil {
		OnError500(w, err, "Can't list points")
		return
	}
	JsonAnswer(w, points)
}
