package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"github.com/gorilla/handlers"
)

var storage Storage
var files Files

func TracksHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	json, e := json.Marshal(storage.getTracks())
	if e != nil {
		onError(w, e, "Json serialize error")
	} else {
		w.Write(json)
	}
}
func TileHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	callback := req.FormValue("callback")
	_, err := NewBbox(req.FormValue("bbox"))
	if err != nil {
		onError(w, err, "Can not parse bbox")
		return
	}
	tracks := storage.getTracks()
	features := TracksToYmaps(tracks)

	w.Write([]byte(callback + "(" + JsonStr(features, "{}") + "); trackList(" + JsonStr(tracks.withoutPath(), "[]") + ");"))
}

func JsonStr(f interface{}, _default string) string {
	bytes, err := json.Marshal(f)
	if err != nil {
		log.Errorf("Can not serialize object %v: %s", f, err.Error())
		return _default
	}
	return string(bytes)
}

func TrackFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	id := vars["id"]

	fileReader, err := files.Get(id)
	if err != nil {
		onError(w, err, "Read data error")
		return
	}

	_, err = io.Copy(w, fileReader)

	if err != nil {
		onError(w, err, "Send error")
	}
}

func onError(w http.ResponseWriter, err error, msg string) {
	errStr := fmt.Sprintf("%s: %v", msg, err)
	log.Errorf(errStr)
	http.Error(w, errStr, http.StatusInternalServerError)
}

func main() {
	log.Infof("Starting wwmap")

	storage = NewPostgresStorage()
	files = DummyFiles{}

	r := mux.NewRouter()
	r.HandleFunc("/tracks", TracksHandler)
	r.HandleFunc("/tile", TileHandler)
	r.HandleFunc("/track-files/{id}", TrackFiles)

	httpStr := fmt.Sprintf(":%d", 7007)
	log.Infof("Starting http server on %s", httpStr)
	http.Handle("/", r)
	err := http.ListenAndServe(httpStr, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
