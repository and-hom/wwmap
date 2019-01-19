package handler

import (
	"encoding/json"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"net/http"
	"strconv"
)

type SystemHandler struct {
	App
	versionJson       []byte
	versionMarshalErr error
}

func CreateSystemHandler(app *App, version string) *SystemHandler {
	var versionJson, versionMarshalErr = json.Marshal(version)
	return &SystemHandler{*app, versionJson, versionMarshalErr,}
}

func (this *SystemHandler) Init() {
	this.Register("/version", HandlerFunctions{Get: this.Version})
	this.Register("/log", HandlerFunctions{Get: this.ForRoles(this.Log, dao.ADMIN)})
}

func (this *SystemHandler) Version(w http.ResponseWriter, req *http.Request) {
	if this.versionMarshalErr != nil {
		OnError500(w, this.versionMarshalErr, "Can not marshal version to json")
	}
	_, err := w.Write(this.versionJson)
	if err != nil {
		OnError500(w, err, "Can not write version to response")
	}
}

func (this *SystemHandler) Log(w http.ResponseWriter, req *http.Request) {
	objectType := req.FormValue("object_type")
	objectId, err := strconv.ParseInt(req.FormValue("object_id"), 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse object id", http.StatusBadRequest)
	}
	entries, err := this.ChangesLogDao.List(objectType, objectId, 100)
	if err != nil {
		OnError500(w, err, "Can not fetch log entries")
	}
	this.JsonAnswer(w, entries)
}
