package handler

import (
	"encoding/json"
	"net/http"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
)

type RefSitesHandler struct { App };

func (this *RefSitesHandler) Init() {
	this.Register("/ref-sites", HandlerFunctions{Get:this.RefSites})
}

func (this *RefSitesHandler) RefSites(w http.ResponseWriter, req *http.Request) {
	JsonResponse(w)

	refs, err := this.RefererStorage.List()
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
