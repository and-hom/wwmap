package handler

import (
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"net/http"
	"strconv"
)

type SystemHandler struct {
	App
	DbVersionDao dao.DbVersionDao
	version      string
}

func CreateSystemHandler(app *App, dbVersionDao dao.DbVersionDao, version string) *SystemHandler {
	return &SystemHandler{*app, dbVersionDao, version}
}

func (this *SystemHandler) Init() {
	this.Register("/version", HandlerFunctions{Get: this.Version})
	this.Register("/db-version", HandlerFunctions{Get: this.DbVersion})
	this.Register("/log", HandlerFunctions{Get: this.ForRoles(this.Log, dao.ADMIN)})
}

func (this *SystemHandler) Version(w http.ResponseWriter, req *http.Request) {
	this.JsonAnswer(w, this.version)
}

func (this *SystemHandler) DbVersion(w http.ResponseWriter, req *http.Request) {
	dbVersion, err := this.DbVersionDao.GetDbVersion()
	if err != nil {
		OnError500(w, err, "Can't select schema version")
		return
	}
	this.JsonAnswer(w, dbVersion)
}

func (this *SystemHandler) Log(w http.ResponseWriter, req *http.Request) {
	objectType := req.FormValue("object_type")
	if objectType == "" {
		lastRows, err := this.ChangesLogDao.ListAll(300)
		if err != nil {
			OnError500(w, err, "Can not fetch log entries")
		}
		this.JsonAnswer(w, lastRows)
		return
	}

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
