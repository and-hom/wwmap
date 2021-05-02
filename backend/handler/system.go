package handler

import (
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"io/ioutil"
	"net/http"
	"regexp"
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
	this.Register("/gav", HandlerFunctions{Get: this.Gav})
}

func (this *SystemHandler) Version(w http.ResponseWriter, req *http.Request) {
	JsonAnswer(w, this.version)
}

func (this *SystemHandler) DbVersion(w http.ResponseWriter, req *http.Request) {
	dbVersion, err := this.DbVersionDao.GetDbVersion()
	if err != nil {
		OnError500(w, err, "Can't select schema version")
		return
	}
	JsonAnswer(w, dbVersion)
}

func (this *SystemHandler) Log(w http.ResponseWriter, req *http.Request) {
	objectType := req.FormValue("object_type")
	if objectType == "" {
		lastRows, err := this.ChangesLogDao.ListAllWithUserInfo(300)
		if err != nil {
			OnError500(w, err, "Can not fetch log entries")
			return
		}
		JsonAnswer(w, lastRows)
		return
	}

	objectId, err := strconv.ParseInt(req.FormValue("object_id"), 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse object id", http.StatusBadRequest)
	}
	entries, err := this.ChangesLogDao.ListWithUserInfo(objectType, objectId, 100)
	if err != nil {
		OnError500(w, err, "Can not fetch log entries")
		return
	}
	JsonAnswer(w, entries)
}

const DEFAULT_GAV = "865"

var gav = ""
var gavRe = regexp.MustCompile("https://khms\\d+.googleapis\\.com/kh\\?v=(\\d+)")

func (this *SystemHandler) Gav(w http.ResponseWriter, req *http.Request) {
	if gav != "" {
		JsonAnswer(w, gav)
		return
	}

	client := http.Client{}
	resp, err := client.Get("https://maps.googleapis.com/maps/api/js")
	if err != nil {
		log.Error("Can't fetch googleapi script:", err)
		JsonAnswer(w, DEFAULT_GAV)
		return
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Can't read googleapi script:", err)
		JsonAnswer(w, DEFAULT_GAV)
		return
	}

	found := gavRe.FindStringSubmatch(string(bytes))
	if found != nil && len(found) > 1 {
		gav = found[1]
		log.Info("Detected google api version is ", gav)
		JsonAnswer(w, gav)
	} else {
		log.Error("Can't find googleapi version")
		JsonAnswer(w, DEFAULT_GAV)
	}
}
