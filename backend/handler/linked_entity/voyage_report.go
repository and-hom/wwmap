package linked_entity

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/backend/handler"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/mux"
	"math"
	"net/http"
	"strconv"
	"time"
)

type VoyageReportHandler struct {
	handler.App
	VoyageReportDao dao.VoyageReportDao
}

func (this *VoyageReportHandler) Create() linkedEntityHanler {
	return linkedEntityHanler{
		this.App,
		this,
		"voyage-report",
		[]dao.Role{dao.ADMIN},
	}
}

func (this *VoyageReportHandler) List(w http.ResponseWriter, r *http.Request) {
	JsonAnswerF(w, func() (i interface{}, err error) {
		withRivers := handler.GetBoolParameter(r, "rivers", false)
		_, allowed, err := CheckRoleAllowed(r, this.UserDao, dao.ADMIN, dao.EDITOR)
		if err != nil {
			log.Error("Can't detect user role ", err)
		}
		return this.VoyageReportDao.List(withRivers, err == nil && allowed)
	}, "Can't list voyageReport records")
}

func (this *VoyageReportHandler) ByRiver(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	JsonAnswerF(w, func() (i interface{}, err error) {
		return this.VoyageReportDao.ByRiver(id, math.MaxInt32)
	}, "Can't list voyageReport records")
}

func (this *VoyageReportHandler) Get(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	voyageReportId, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.writeVoyageReport(voyageReportId, w)
}

func (this *VoyageReportHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	voyageReport := dao.VoyageReport{}
	body, err := DecodeJsonBody(r, &voyageReport)
	if err != nil {
		OnError500(w, err, "Can not parse json from request body: "+body)
		return
	}

	var id int64
	var logType dao.ChangesLogEntryType
	if voyageReport.Id > 0 {
		err = this.VoyageReportDao.Update(voyageReport)
		id = voyageReport.Id
		logType = dao.ENTRY_TYPE_MODIFY
	} else {
		id, err = this.VoyageReportDao.Insert(voyageReport)
		voyageReport.DateModified = time.Now()
		logType = dao.ENTRY_TYPE_CREATE
	}
	if err != nil {
		OnError500(w, err, "Can not save VoyageReport: "+body)
		return
	}

	this.writeVoyageReport(id, w)

	this.LogUserEvent(r, handler.VOYAGE_REPORT_LOG_ENTRY_TYPE, id, logType, voyageReport.Url)
}

func (this *VoyageReportHandler) writeVoyageReport(voyageReportId int64, w http.ResponseWriter) {
	voyageReport, found, err := this.VoyageReportDao.Find(voyageReportId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get voyageReport %d", voyageReportId))
		return
	}
	if !found {
		OnError(w, nil, fmt.Sprintf("VoyageReport with id %d not found", voyageReportId), http.StatusNotFound)
		return
	}
	JsonAnswer(w, voyageReport)
}

func (this *VoyageReportHandler) Delete(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	voyageReportIdStr := pathParams["id"]
	voyageReportId, err := strconv.ParseInt(voyageReportIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	voyageReport, found, err := this.VoyageReportDao.Find(voyageReportId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not select voyageReport by id: %d", voyageReportId))
		return
	}
	if !found {
		OnError(w, err, fmt.Sprintf("VoyageReport with id %d not found", voyageReportId), http.StatusNotFound)
		return
	}

	err = this.VoyageReportDao.Remove(voyageReportId)

	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not remove voyageReport by id: %d", voyageReportId))
		return
	}

	this.LogUserEvent(r, handler.VOYAGE_REPORT_LOG_ENTRY_TYPE, voyageReportId, dao.ENTRY_TYPE_DELETE, voyageReport.Url)
}

func (this *VoyageReportHandler) UndoDelete(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	voyageReportIdStr := pathParams["id"]
	voyageReportId, err := strconv.ParseInt(voyageReportIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	voyageReport, found, err := this.VoyageReportDao.Find(voyageReportId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not select voyageReport by id: %d", voyageReportId))
		return
	}
	if !found {
		OnError(w, err, fmt.Sprintf("VoyageReport with id %d not found", voyageReportId), http.StatusNotFound)
		return
	}

	err = this.VoyageReportDao.UndoRemove(voyageReportId)

	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not undo delete of voyageReport by id: %d", voyageReportId))
		return
	}

	this.LogUserEvent(r, handler.VOYAGE_REPORT_LOG_ENTRY_TYPE, voyageReportId, dao.ENTRY_TYPE_MODIFY, voyageReport.Url)
}
