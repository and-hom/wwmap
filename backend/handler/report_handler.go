package handler

import (
	"net/http"
	"fmt"
	"strconv"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
	"time"
)

type ReportHandler struct {
	App
}

func (this *ReportHandler) Init() {
	this.Register("/report", HandlerFunctions{Post:this.AddReport, Put:this.AddReport})
}

func (this *ReportHandler) AddReport(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	objectIdStr := r.FormValue("object_id")
	objectTitle := r.FormValue("object_title")
	comment := r.FormValue("comment")
	objectId, err := strconv.ParseInt(objectIdStr, 10, 64)
	if err != nil {
		OnError(w, err, fmt.Sprintf("Can not parse object id: %s", objectIdStr), 400)
		return
	}
	err = this.NotificationHelper.SendToRole(Notification{
		IdTitle: IdTitle{Title:title},
		Object: IdTitle{Id:objectId, Title:objectTitle},
		Comment: comment,
		Classifier:"report",
		SendBefore:time.Now().Add(2 * time.Hour), // wait 2 hours for more messages
	}, ADMIN)
	if err != nil {
		OnError500(w, err, "Can not add report")
		return
	}
}