package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"fmt"
)

var storage Storage

func TracksHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	json, e := json.Marshal(storage.getTracks())
	if e != nil {
		log.Errorf("Json serialize error: %v", e)
		http.Error(w, "Unknown error", http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

func main() {
	log.Infof("Starting wwmap")

	storage = DummyStorage{}

	r := mux.NewRouter()
	r.HandleFunc("/tracks", TracksHandler)

	httpStr := fmt.Sprintf(":%d", 7007)
	log.Infof("Starting http server on %s", httpStr)
	http.Handle("/", r)
	err := http.ListenAndServe(httpStr, nil)
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
