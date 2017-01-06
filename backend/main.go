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
	"errors"
	"strconv"
)

var storage Storage
var files Files

func TileHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	callback := req.FormValue("callback")
	bbox, err := NewBbox(req.FormValue("bbox"))
	if err != nil {
		onError(w, err, "Can not parse bbox")
		return
	}
	tracks := storage.getTracks(bbox)
	features := TracksToYmaps(tracks)
	log.Infof("Found %d", len(tracks))

	w.Write([]byte(callback + "(" + JsonStr(features, "{}") + ");"))
}

func SingleTrackTileHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	callback := req.FormValue("callback")
	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id")
		return
	}
	track := storage.getTrack(id)
	features := TracksToYmaps(track)
	w.Write([]byte(callback + "(" + JsonStr(features, "{}") + ",{strokeWidth:5});"))
}

func SingleTrackBoundsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id")
		return
	}
	tracks := storage.getTrack(id)
	if len(tracks) > 0 {
		bytes, err := json.Marshal(tracks[0].Bounds())
		if err != nil {
			onError(w, err, "Can not serialize track")
			return
		}
		w.Write(bytes)
	} else {
		onError(w, err, "Track not found")
	}
}

func TracksHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	bbox, err := NewBbox(req.FormValue("bbox"))
	if err != nil {
		onError(w, err, "Can not parse bbox")
		return
	}
	tracks := storage.getTracks(bbox)
	log.Infof("Found %d", len(tracks))

	w.Write([]byte(JsonStr(tracks.withoutPath(), "[]")))
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

func UploadTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		onError(w, err, "Can not parse form")
		return
	}

	//title := r.FormValue("title")
	redirectTo := r.FormValue("redirect_to")
	cookieDomain := r.FormValue("cookie_domain")
	cookiePath := r.FormValue("cookie_path")

	file, _, err := r.FormFile("geo_file")
	if err != nil {
		onError(w, err, "Can not read form file")
		return
	}

	parser, err := geoParser(file)
	if err != nil {
		onError(w, err, "Can not parse geo data")
		return
	}
	tracks, err := parser.getTracks()
	if err != nil {
		onError(w, err, "Bad geo data symantics")
		return
	}
	err = storage.insert(tracks...)
	if err != nil {
		onError(w, err, "Can not insert tracks")
		return
	}

	if len(tracks) > 0 && len(tracks[0].Path) > 0 {
		jsonBytes, err := tracks[0].Path[0].MarshalJSON()
		if err == nil {
			viewPositionCookie := http.Cookie{
				Name:"last-map-pos",
				Value:string(jsonBytes),
				Domain:cookieDomain,
				Path:cookiePath,
			}
			http.SetCookie(w, &viewPositionCookie)
		} else {
			log.Warnf("Can not marshal point %v", tracks[0].Path[0])
		}
	}
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

func onError(w http.ResponseWriter, err error, msg string) {
	errStr := fmt.Sprintf("%s: %v", msg, err)
	log.Errorf(errStr)
	http.Error(w, errStr, http.StatusInternalServerError)
}

func geoParser(r io.ReadSeeker) (GeoParser, error) {
	gpxParser, err := InitGpxParser(r)
	if err == nil {
		return gpxParser, nil
	}
	log.Warn(err)
	r.Seek(0, 0)
	kmlParser, err := InitKmlParser(r)
	if err == nil {
		return kmlParser, nil
	}
	log.Warn(err)
	return nil, errors.New("Can not find valid parser for this format!")
}

func main() {
	log.Infof("Starting wwmap")

	storage = NewPostgresStorage()
	files = DummyFiles{}

	r := mux.NewRouter()
	r.HandleFunc("/tile", TileHandler)
	r.HandleFunc("/single-track-tile", SingleTrackTileHandler)
	r.HandleFunc("/single-track-bounds", SingleTrackBoundsHandler)
	r.HandleFunc("/tracks", TracksHandler)
	r.HandleFunc("/track-files/{id}", TrackFiles)
	r.HandleFunc("/upload-track", UploadTrack).Methods("POST")

	httpStr := fmt.Sprintf(":%d", 7007)
	log.Infof("Starting http server on %s", httpStr)
	http.Handle("/", r)
	err := http.ListenAndServe(httpStr, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
