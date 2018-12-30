package handler

import (
	"encoding/json"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"net/http"
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
