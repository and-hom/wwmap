package handler

import (
	"net/http"
	"fmt"
	"strconv"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/gorilla/mux"
)

type ReportHandler struct {
	App
}

func (this *ReportHandler) Init(r *mux.Router) {
	this.Register(r, "/report", HandlerFunctions{Post:this.AddReport, Put:this.AddReport})
}

func (this *ReportHandler) AddReport(w http.ResponseWriter, r *http.Request) {
	comment := r.FormValue("comment")
	objectIdStr := r.FormValue("object_id")
	objectId, err := strconv.ParseInt(objectIdStr, 10, 64)
	if err != nil {
		OnError(w, err, fmt.Sprintf("Can not parse object id: %s", objectIdStr), 400)
		return
	}
	err = this.ReportDao.AddReport(Report{
		ObjectId: objectId,
		Comment: comment,
	})
	if err != nil {
		OnError500(w, err, "Can not add report")
		return
	}
}