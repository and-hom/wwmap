package handler

import (
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/handler"
	http2 "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type TransferHandler struct {
	App
	TransferDao dao.TransferDao
}

func (this *TransferHandler) Init() {
	this.Register("/transfer", handler.HandlerFunctions{
		Get:  this.List,
		Post: this.ForRoles(this.Upsert, dao.ADMIN, dao.EDITOR),
		Put:  this.ForRoles(this.Upsert, dao.ADMIN, dao.EDITOR),
	})
	this.Register("/transfer-full", handler.HandlerFunctions{
		Get: this.ListFull,
	})
	this.Register("/transfer/{id}", handler.HandlerFunctions{
		Delete: this.ForRoles(this.Delete, dao.ADMIN, dao.EDITOR),
	})
	this.Register("/transfer/river/{id}", handler.HandlerFunctions{
		Get: this.ByRiver,
	})
}

func (this *TransferHandler) List(w http.ResponseWriter, req *http.Request) {
	this.JsonAnswerF(w, func() (i interface{}, err error) {
		return this.TransferDao.List()
	}, "Can't list transfer records")
}

func (this *TransferHandler) ListFull(w http.ResponseWriter, req *http.Request) {
	this.JsonAnswerF(w, func() (i interface{}, err error) {
		return this.TransferDao.ListFull()
	}, "Can't list transfer records")
}

func (this *TransferHandler) ByRiver(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		http2.OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	this.JsonAnswerF(w, func() (i interface{}, err error) {
		return this.TransferDao.ByRiver(id)
	}, "Can't list transfer records")
}

func (this *TransferHandler) Upsert(w http.ResponseWriter, req *http.Request) {
	transfer := dao.TransferFull{}
	body, err := handler.DecodeJsonBody(req, &transfer)
	if err != nil {
		http2.OnError500(w, err, "Can not parse json from request body: "+body)
		return
	}

	if transfer.Id > 0 {
		if err := this.TransferDao.Update(transfer); err != nil {
			http2.OnError500(w, err, "Can't update")
			return
		}
	} else {
		if _, err := this.TransferDao.Insert(transfer); err != nil {
			http2.OnError500(w, err, "Can't insert")
			return
		}
	}
	this.JsonAnswer(w, true)
}

func (this *TransferHandler) Delete(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		http2.OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	if err := this.TransferDao.Remove(id); err != nil {
		http2.OnError500(w, err, fmt.Sprintf("Can't delete transfer with id %d", id))
		return
	}
}