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
	"time"
	"io/ioutil"
)

var storage Storage

func TileHandler(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET, OPTIONS")

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
	corsHeaders(w, "GET, OPTIONS")

	callback := req.FormValue("callback")
	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	track := storage.getTrack(id)
	features := TracksToYmaps(track)

	w.Write(JsonpAnswer(callback, features, "{}"))
}

func SingleTrackBoundsHandler(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET, OPTIONS")

	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	tracks := storage.getTrack(id)
	if len(tracks) > 0 {
		bytes, err := json.Marshal(tracks[0].Bounds(true))
		if err != nil {
			onError500(w, err, "Can not serialize track")
			return
		}
		w.Write(bytes)
	} else {
		onError500(w, err, "Track not found")
	}
}

func TrackEditorPageHandler(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET, OPTIONS")

	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	tracks := storage.getTrack(id)
	if len(tracks) > 0 {
		track := tracks[0]
		bytes, err := json.Marshal(TrackEditorPage{
			Title:track.Title,
			Type: track.Type,
			TrackBounds: track.Bounds(true),
			EventPoints: track.Points,
		})
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
	corsHeaders(w, "GET, OPTIONS")

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
						RenderedGeometry: NewYRectangleInt([][]int{
							{tilePoint.x - 15, tilePoint.y - 15, },
							{tilePoint.x + 15, tilePoint.y + 15, },
						}),
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
	corsHeaders(w, "GET")

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

func GetTrack(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")
	// for cors only
}

func EditTrack(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")
	err := r.ParseForm()
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	track, err := parseTrackForm(w, r)
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}
	track.Id = id

	err = storage.UpdateTrack(track)
	if err != nil {
		onError500(w, err, "Can not edit track	")
		return
	}
	writeTrackToResponse(id, w)
}

func writeTrackToResponse(id int64, w http.ResponseWriter) {
	track := Track{}
	found, err := storage.FindTrack(id, &track)
	if err != nil {
		onError500(w, err, "Can not find")
		return
	}
	if !found {
		onError(w, fmt.Errorf("Point with id %d does not exist", id), "Not found", http.StatusNotFound)
	}
	bytes, err := json.Marshal(track)
	if err != nil {
		onError500(w, err, "Can not marshal")
		return
	}
	w.Write(bytes)
}

func parseTrackForm(w http.ResponseWriter, r *http.Request) (Track, error) {
	title := r.FormValue("title")
	tType, err := parseTrackType(r.FormValue("type"))
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return Track{}, err
	}
	track := Track{
		Type:tType,
		Title:title,
	}

	return track, nil
}

func GetPoint(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	writePointToResponse(id, w)
}

func writePointToResponse(id int64, w http.ResponseWriter) {
	eventPoint := EventPoint{}
	found, err := storage.FindEventPoint(id, &eventPoint)
	if err != nil {
		onError500(w, err, "Can not find")
		return
	}
	if !found {
		onError(w, fmt.Errorf("Point with id %d does not exist", id), "Not found", http.StatusNotFound)
	}
	bytes, err := json.Marshal(eventPoint)
	if err != nil {
		onError500(w, err, "Can not marshal")
		return
	}
	w.Write(bytes)
}

func DelPoint(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")
	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	err = storage.DeleteEventPoint(id)
	if err != nil {
		onError500(w, err, "Can not delete")
		return
	}
}

func EditPoint(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")
	err := r.ParseForm()
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	eventPoint, _, err := parsePointForm(w, r)
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}
	eventPoint.Id = id

	err = storage.UpdateEventPoint(eventPoint)
	if err != nil {
		onError500(w, err, "Can not edit point	")
		return
	}
	writePointToResponse(id, w)
}

func AddPoint(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")
	err := r.ParseForm()
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}

	eventPoint, trackId, err := parsePointForm(w, r)
	if err != nil {
		return
	}

	id, err := storage.AddEventPoint(trackId, eventPoint)

	if err != nil {
		onError500(w, err, "Can not insert")
		return
	}

	w.Write([]byte(strconv.FormatInt(id, 10)))
}

func parsePointForm(w http.ResponseWriter, r *http.Request) (EventPoint, int64, error) {
	trackId, err := strconv.ParseInt(r.FormValue("track_id"), 10, 64)
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return EventPoint{}, 0, err
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	pType, err := parseEventPointType(r.FormValue("type"))
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return EventPoint{}, 0, err
	}
	point := Point{};
	err = json.Unmarshal([]byte(r.FormValue("point")), &point)
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return EventPoint{}, 0, err
	}

	eventPoint := EventPoint{
		Type:pType,
		Title:title,
		Content:content,
		Point:point,
		Time:JSONTime(time.Now()),
	}

	return eventPoint, trackId, nil
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
	err = storage.AddTracks(tracks...)
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

func PictureMetadataHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST")

	requestBody := r.Body
	defer requestBody.Close()

	imgUrl, err := ioutil.ReadAll(requestBody)
	if err != nil {
		onError(w, err, "Can not read request body", http.StatusBadRequest)
		return
	}

	imgResp, err := http.Get(string(imgUrl))
	if err != nil {
		onError(w, err, "Can not fetch image", 422)
		return
	}

	defer imgResp.Body.Close()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "img")
	if err != nil {
		onError500(w, err, "Can not create temp file")
		return
	}
	defer CloseAndRemove(tmpFile)

	_, err = io.Copy(tmpFile, imgResp.Body)
	if err != nil {
		onError500(w, err, "Can not fetch image from server: " + string(imgUrl))
		return
	}
	_, err = tmpFile.Seek(0, os.SEEK_SET)
	if err != nil {
		onError500(w, err, "Can not seek on img file")
		return
	}

	imgData, err := GetImgProperties(tmpFile)
	if err != nil {
		onError500(w, err, "Can not get img properties")
		return
	}

	w.Write([]byte(JsonStr(imgData, "{}")))
}

func CloseAndRemove(f *os.File) {
	f.Close()
	os.Remove(f.Name())
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

func corsHeaders(w http.ResponseWriter, methods string) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", methods)
	w.Header().Add("Access-Control-Allow-Headers", "origin, x-csrftoken, content-type, accept")
}

func main() {
	log.Infof("Starting wwmap")

	storage = NewPostgresStorage()

	r := mux.NewRouter()
	r.HandleFunc("/tile", TileHandler)
	r.HandleFunc("/single-track-tile", SingleTrackTileHandler)
	r.HandleFunc("/single-track-bounds", SingleTrackBoundsHandler)
	r.HandleFunc("/track-editor-page", TrackEditorPageHandler)
	r.HandleFunc("/track-active-areas/{id:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}/{z:[0-9]+}", TrackPointsToClickHandler)
	r.HandleFunc("/tracks", TracksHandler)

	r.HandleFunc("/track/{id}", EditTrack).Methods("PUT")
	r.HandleFunc("/track/{id}", GetTrack).Methods("GET", "OPTIONS")

	r.HandleFunc("/upload-track", UploadTrack).Methods("POST")

	r.HandleFunc("/point", AddPoint).Methods("POST")
	r.HandleFunc("/point/{id}", EditPoint).Methods("PUT")
	r.HandleFunc("/point/{id}", DelPoint).Methods("DELETE")
	r.HandleFunc("/point/{id}", GetPoint).Methods("OPTIONS", "GET")
	r.HandleFunc("/picture-metadata", PictureMetadataHandler).Methods("POST")

	httpStr := fmt.Sprintf(":%d", 7007)
	log.Infof("Starting http server on %s", httpStr)
	http.Handle("/", r)
	err := http.ListenAndServe(httpStr, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
