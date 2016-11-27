package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"fmt"
	"io"
)

var storage Storage
var files Files

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

func TrackFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	id := vars["id"]

	fileReader, err := files.Get(id)
	if err != nil {
		log.Errorf("Read data error: %v", err)
		http.Error(w, "Unknown error", http.StatusInternalServerError)
	}

	_, err = io.Copy(w, fileReader)

	if err != nil {
		log.Errorf("Send error: %v", err)
		http.Error(w, "Unknown error", http.StatusInternalServerError)
	}
}

func main() {
	log.Infof("Starting wwmap")

	storage = DummyStorage{}
	files = DummyFiles{}

	r := mux.NewRouter()
	r.HandleFunc("/tracks", TracksHandler)
	r.HandleFunc("/track-files/{id}", TrackFiles)

	httpStr := fmt.Sprintf(":%d", 7007)
	log.Infof("Starting http server on %s", httpStr)
	http.Handle("/", r)
	err := http.ListenAndServe(httpStr, nil)
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
