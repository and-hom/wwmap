package linked_entity

import (
	"fmt"
	"github.com/and-hom/wwmap/backend/handler"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type TransferHandler struct {
	handler.App
	TransferDao dao.TransferDao
}

func (this *TransferHandler) Create() linkedEntityHanler {
	return linkedEntityHanler{
		this.App,
		this,
		"transfer",
	}
}

func (this *TransferHandler) List(w http.ResponseWriter, r *http.Request) {
	JsonAnswerF(w, func() (i interface{}, err error) {
		withRivers := handler.GetBoolParameter(r, "rivers", false)
		return this.TransferDao.List(withRivers)
	}, "Can't list transfer records")
}

func (this *TransferHandler) Get(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.writeTransfer(id, w)
}

func (this *TransferHandler) ByRiver(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	JsonAnswerF(w, func() (i interface{}, err error) {
		return this.TransferDao.ByRiver(id)
	}, "Can't list transfer records")
}

func (this *TransferHandler) Upsert(w http.ResponseWriter, req *http.Request) {
	transfer := dao.Transfer{}
	body, err := DecodeJsonBody(req, &transfer)
	if err != nil {
		OnError500(w, err, "Can not parse json from request body: "+body)
		return
	}

	if transfer.Id > 0 {
		if err := this.TransferDao.Update(transfer); err != nil {
			OnError500(w, err, "Can't update")
			return
		}
	} else {
		if transfer.Id, err = this.TransferDao.Insert(transfer); err != nil {
			OnError500(w, err, "Can't insert")
			return
		}
	}
	JsonAnswer(w, transfer)
}

func (this *TransferHandler) Delete(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	if err := this.TransferDao.Remove(id); err != nil {
		OnError500(w, err, fmt.Sprintf("Can't delete transfer with id %d", id))
		return
	}
}

func (this *TransferHandler) writeTransfer(id int64, w http.ResponseWriter) {
	camp, found, err := this.TransferDao.Find(id)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not get camp %d", id))
		return
	}
	if !found {
		OnError(w, nil, fmt.Sprintf("Transfer with id %d not found", id), http.StatusNotFound)
		return
	}
	JsonAnswer(w, camp)
}
