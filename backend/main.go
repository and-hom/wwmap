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
		onError500(w, err, "Can not parse bbox")
		return
	}
	tracks := storage.getTracks(bbox)
	features := TracksToYmaps(tracks)
	log.Infof("Found %d", len(tracks))

	w.Write(JsonpAnswer(callback, features, "{}"))
}

func SingleTrackTileHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	callback := req.FormValue("callback")
	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		onError500(w, err, "Can not parse id")
		return
	}
	track := storage.getTrack(id)
	features := TracksToYmaps(track)

	w.Write(JsonpAnswer(callback, features, "{}"))
}

func SingleTrackBoundsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		onError500(w, err, "Can not parse id")
		return
	}
	tracks := storage.getTrack(id)
	if len(tracks) > 0 {
		bytes, err := json.Marshal(tracks[0].Bounds())
		if err != nil {
			onError500(w, err, "Can not serialize track")
			return
		}
		w.Write(bytes)
	} else {
		onError500(w, err, "Track not found")
	}
}

func TrackPointsToClickHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	callback := req.FormValue("callback")
	pathParams := mux.Vars(req)

	x, err := strconv.Atoi(pathParams["x"])
	if err != nil {
		onError500(w, err, "Can not parse x")
		return
	}

	y, err := strconv.Atoi(pathParams["y"])
	if err != nil {
		onError500(w, err, "Can not parse y")
		return
	}

	z, err := strconv.ParseUint(pathParams["z"], 10, 32)
	if err != nil {
		onError500(w, err, "Can not parse z")
		return
	}

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		onError500(w, err, "Can not parse id")
		return
	}

	tracks := storage.getTrack(id)
	var t Track
	if len(tracks) > 0 {
		t = tracks[0]
	} else {
		onError(w, errors.New("Not found"), fmt.Sprintf("Can not find track id = %d", id), http.StatusNotFound)
		return
	}
	track := t

	features := make([]Feature, 0)
	for i, point := range track.Path {
		tilePoint, tileX, tileY := toTileCoords(uint32(z), point)
		if x == tileX && y == tileY {
			features = append(features, Feature{
				Properties:FeatureProperties{
					HotspotMetaData : HotspotMetaData{
						Id:int64(i * 30 + int(z)),
						RenderedGeometry: NewYPolygonInt([][][]int{{
							{tilePoint.x - 15, tilePoint.y - 15, },
							{tilePoint.x + 15, tilePoint.y - 15, },
							{tilePoint.x + 15, tilePoint.y + 15, },
							{tilePoint.x - 15, tilePoint.y + 15, },
							{tilePoint.x - 15, tilePoint.y - 15, },
						}}),
					},
				},
				Type:"Feature",
			})
		}
	}
	featureCollection := FeatureCollectionWrapper{
		FeatureCollection: FeatureCollection{
			Features:features,
			Type:"FeatureCollection",
		},
	}

	w.Write(JsonpAnswer(callback, featureCollection, "{}"))
}

func TracksHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	bbox, err := NewBbox(req.FormValue("bbox"))
	if err != nil {
		onError500(w, err, "Can not parse bbox")
		return
	}
	tracks := storage.getTracks(bbox)
	log.Infof("Found %d", len(tracks))

	w.Write([]byte(JsonStr(tracks.withoutPath(), "[]")))
}

func JsonpAnswer(callback string, object interface{}, _default string) []byte {
	return []byte(callback + "(" + JsonStr(object, _default) + ");")
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
		onError500(w, err, "Read data error")
		return
	}

	_, err = io.Copy(w, fileReader)

	if err != nil {
		onError500(w, err, "Send error")
	}
}

func UploadTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		onError500(w, err, "Can not parse form")
		return
	}

	//title := r.FormValue("title")
	redirectTo := r.FormValue("redirect_to")
	cookieDomain := r.FormValue("cookie_domain")
	cookiePath := r.FormValue("cookie_path")

	file, _, err := r.FormFile("geo_file")
	if err != nil {
		onError500(w, err, "Can not read form file")
		return
	}

	parser, err := geoParser(file)
	if err != nil {
		onError500(w, err, "Can not parse geo data")
		return
	}
	tracks, err := parser.getTracks()
	if err != nil {
		onError500(w, err, "Bad geo data symantics")
		return
	}
	err = storage.insert(tracks...)
	if err != nil {
		onError500(w, err, "Can not insert tracks")
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

func onError(w http.ResponseWriter, err error, msg string, statusCode int) {
	errStr := fmt.Sprintf("%s: %v", msg, err)
	log.Errorf(errStr)
	http.Error(w, errStr, statusCode)
}

func onError500(w http.ResponseWriter, err error, msg string) {
	onError(w, err, msg, http.StatusInternalServerError)
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
	r.HandleFunc("/track-active-areas/{id:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}/{z:[0-9]+}", TrackPointsToClickHandler)
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
