package main

import (
	"encoding/json"
	"net/http"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/mux"
)

type RefSitesHandler struct{ Handler };

func (this *RefSitesHandler) Init(r *mux.Router) {
	this.Register(r, "/ref-sites", HandlerFunctions{get:this.RefSites})
}

func (this *RefSitesHandler) RefSites(w http.ResponseWriter, req *http.Request) {
	JsonResponse(w)

	refs, err := this.refererStorage.List()
	if err != nil {
		OnError500(w, err, "Can not list referers")
		return
	}
	bytes, err := json.Marshal(refs)
	if err != nil {
		OnError500(w, err, "Can not marshal json")
		return
	}
	w.Write(bytes)
}
