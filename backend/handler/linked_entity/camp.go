package linked_entity

import (
	"fmt"
	"github.com/and-hom/wwmap/backend/handler"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/ptrv/go-gpx"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type CampHandler struct {
	handler.App
	CampDao dao.CampDao
}

func (this *CampHandler) Create() linkedEntityHanler {
	return linkedEntityHanler{
		this.App,
		this,
		"camp",
		[]dao.Role{dao.ADMIN, dao.EDITOR},
	}
}

func (this *CampHandler) List(w http.ResponseWriter, r *http.Request) {
	JsonAnswerF(w, func() (i interface{}, err error) {
		withRivers := handler.GetBoolParameter(r, "rivers", false)
		return this.CampDao.List(withRivers)
	}, "Can't list camp records")
}

func (this *CampHandler) ByRiver(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	JsonAnswerF(w, func() (i interface{}, err error) {
		return this.CampDao.ByRiver(id)
	}, "Can't list camp records")
}

func (this *CampHandler) Get(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	campId, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.writeCamp(campId, w)
}

func (this *CampHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	camp := dao.Camp{}
	body, err := DecodeJsonBody(r, &camp)
	if err != nil {
		OnError500(w, err, "Can not parse json from request body: "+body)
		return
	}

	var id int64
	var logType dao.ChangesLogEntryType
	if camp.Id > 0 {
		err = this.CampDao.Update(camp)
		id = camp.Id
		logType = dao.ENTRY_TYPE_MODIFY
	} else {
		id, err = this.CampDao.Insert(camp)
		logType = dao.ENTRY_TYPE_CREATE
	}
	if err != nil {
		OnError500(w, err, "Can not save camp: "+body)
		return
	}

	this.writeCamp(id, w)

	this.LogUserEvent(r, handler.CAMP_LOG_ENTRY_TYPE, id, logType, camp.Title)
}

func (this *CampHandler) writeCamp(campId int64, w http.ResponseWriter) {
	camp, found, err := this.CampDao.Find(campId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get camp %d", campId))
		return
	}
	if !found {
		OnError(w, nil, fmt.Sprintf("Camp with id %d not found", campId), http.StatusNotFound)
		return
	}
	JsonAnswer(w, camp)
}

func (this *CampHandler) UndoDelete(w http.ResponseWriter, r *http.Request) {
	OnError(w, nil, "Not implemented", http.StatusNotImplemented)
}

func (this *CampHandler) Delete(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	campIdStr := pathParams["id"]
	campId, err := strconv.ParseInt(campIdStr, 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	camp, found, err := this.CampDao.Find(campId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not select camp by id: %d", campId))
		return
	}
	if !found {
		OnError(w, err, fmt.Sprintf("Camp with id %d not found", campId), http.StatusNotFound)
		return
	}

	err = this.CampDao.Remove(campId, nil)

	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not remove camp by id: %d", campId))
		return
	}

	this.LogUserEvent(r, handler.CAMP_LOG_ENTRY_TYPE, campId, dao.ENTRY_TYPE_DELETE, camp.Title)
}

func (this *CampHandler) DownloadGpxForRiver(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	riverId, err := strconv.ParseInt(pathParams["riverId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	trStr := req.FormValue("tr")
	transliterate, err := strconv.ParseBool(trStr)
	if err != nil {
		transliterate = trStr != ""
	}

	camps, err := this.CampDao.ByRiver(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not fetch camps for river %d", riverId))
		return
	}
	if len(camps) == 0 {
		OnError(w, nil, fmt.Sprintf("No camps found for river %d", riverId), http.StatusNotFound)
		return
	}

	river, err := this.RiverDao.Find(riverId)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not fetch river %d", riverId))
		return
	}

	filename := river.Title
	if transliterate {
		filename = "Camp " + util.CyrillicToTranslit(filename)
	} else {
		filename = "Стоянки " + filename
	}

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.gpx\"", filename))
	w.Header().Add("Content-Type", "application/gpx+xml")

	waypoints := make([]gpx.Wpt, 0, len(camps))
	for i := 0; i < len(camps); i++ {
		camp := camps[i]

		title := camp.Title
		if title =="" {
			title = fmt.Sprintf("%d", camp.Id)
		}
		description := camp.Description

		if transliterate {
			title = util.CyrillicToTranslit(title)
			description = util.CyrillicToTranslit(description)
		}
		description = util.TrimToLengthWithTrailingDots(description, 1023)

		waypoints = append(waypoints, gpx.Wpt{
				Lat:  camp.Point.Lat,
				Lon:  camp.Point.Lon,
				Cmt:  description,
				Name: title,
			})

	}
	gpxData := gpx.Gpx{
		Waypoints: waypoints,
		Creator:   "wwmap",
	}

	_, err = w.Write(gpxData.ToXML())
	if err!=nil {
		OnError500(w, err, fmt.Sprintf("Can't write camp GPX for river %d", riverId))
	}
}
