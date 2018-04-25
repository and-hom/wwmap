package main

import (
	"net/http"
	"fmt"
	"strconv"
	. "github.com/and-hom/wwmap/lib/dao"
)

type ReportHandler struct {
	Handler
}

func (this *ReportHandler) AddReport(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST")

	comment := r.FormValue("comment")
	objectIdStr := r.FormValue("object_id")
	objectId, err := strconv.ParseInt(objectIdStr, 10, 64)
	if err != nil {
		this.onError(w, err, fmt.Sprintf("Can not parse object id: %s", objectIdStr), 400)
		return
	}
	err = this.reportDao.AddReport(Report{
		ObjectId: objectId,
		Comment: comment,
	})
	if err != nil {
		this.onError500(w, err, "Can not add report")
		return
	}
}