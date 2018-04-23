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
	. "github.com/and-hom/wwmap/backend/dao"
	. "github.com/and-hom/wwmap/backend/geo"
	. "github.com/and-hom/wwmap/backend/geoparser"
	"github.com/and-hom/wwmap/config"
	"github.com/and-hom/wwmap/backend/model"
	gpx "github.com/ptrv/go-gpx"
)

var storage Storage
var clusterMaker ClusterMaker

func TileHandler(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET, OPTIONS")

	callback := req.FormValue("callback")
	bbox, err := NewBbox(req.FormValue("bbox"))
	if err != nil {
		onError500(w, err, "Can not parse bbox")
		return
	}
	tracks := storage.ListTracks(bbox)
	points := storage.ListPoints(bbox)

	featureCollection := MkFeatureCollection(append(pointsToYmaps(points), tracksToYmaps(tracks)...))
	log.Infof("Found %d", len(featureCollection.Features))

	w.Write(JsonpAnswer(callback, featureCollection, "{}"))
}

func TileWhiteWaterHandler(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET, OPTIONS")

	callback := req.FormValue("callback")
	bbox, err := NewBbox(req.FormValue("bbox"))
	if err != nil {
		onError(w, err, "Can not parse bbox", http.StatusBadRequest)
		return
	}
	points := storage.ListWhiteWaterPoints(bbox)
	waterwayTitles, err := storage.ListWaterWayTitles(bbox, 1024)
	if err != nil {
		onError500(w, err, "Can not select waterways")
		return
	}
	waterwayTitlesMap := make(map[int64]string)
	for _, ww := range waterwayTitles {
		waterwayTitlesMap[ww.Id] = ww.Title
	}

	featureCollection := MkFeatureCollection(whiteWaterPointsToYmaps(points, bbox.Width(), bbox.Height(), waterwayTitlesMap))
	log.Infof("Found %d", len(featureCollection.Features))

	w.Write(JsonpAnswer(callback, featureCollection, "{}"))
}

func GetNearestRivers(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "GET")
	lat_s := r.FormValue("lat")
	lat, err := strconv.ParseFloat(lat_s, 64)
	if err != nil {
		onError(w, err, fmt.Sprintf("Can not parse lat parameter: %s", lat_s), 400)
		return
	}
	lon_s := r.FormValue("lon")
	lon, err := strconv.ParseFloat(lon_s, 64)
	if err != nil {
		onError(w, err, fmt.Sprintf("Can not parse lon parameter: %s", lon_s), 400)
		return
	}
	point := Point{Lat:lat, Lon:lon}
	waterways, err := storage.NearestWaterWays(point, 5)
	if err != nil {
		onError500(w, err, "Can not select rivers")
		return
	}
	w.Write([]byte(JsonStr(waterways, "[]")))
}

func GetVisibleRivers(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET")

	bbox, err := NewBbox(req.FormValue("bbox"))
	if err != nil {
		onError500(w, err, "Can not parse bbox")
		return
	}

	waterways, err := storage.ListWaterWayTitles(bbox, 30)
	if err != nil {
		onError500(w, err, "Can not select rivers")
		return
	}
	w.Write([]byte(JsonStr(waterways, "[]")))
}

func DownloadGpx(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET")
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	waterway, err := storage.WaterWayById(id)
	if err != nil {
		onError(w, err, fmt.Sprintf("Can not find river with id %d", id), http.StatusNotFound)
		return
	}

	whitewaterPoints := storage.ListWhiteWaterPointsByRiver(id)
	waypoints := make([]gpx.Wpt, len(whitewaterPoints))
	for i := 0; i < len(whitewaterPoints); i++ {
		whitewaterPoint := whitewaterPoints[i]
		waypoints[i] = gpx.Wpt{
			Lat: whitewaterPoint.Point.Lat,
			Lon: whitewaterPoint.Point.Lon,
			Name: whitewaterPoint.Title,
			Cmt: whitewaterPoint.Comment,
		}
	}
	gpxData := gpx.Gpx{
		Waypoints: waypoints,
	}
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.gpx\"", waterway.Title))
	w.Header().Add("Content-Type", "application/gpx+xml")

	xmlBytes := gpxData.ToXML()
	w.Write(xmlBytes)
}

func AddWhiteWaterPoints(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")
	err := r.ParseForm()
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}

	wwPoints, err := parseWhiteWaterPointsForm(w, r)
	if err != nil {
		onError500(w, err, "Can not read request")
		return
	}

	err = storage.AddWhiteWaterPoints(wwPoints...)
	fmt.Printf("%v\n", wwPoints)

	if err != nil {
		onError500(w, err, "Can not insert")
		return
	}
}

func parseWhiteWaterPointsForm(w http.ResponseWriter, r *http.Request) ([]WhiteWaterPoint, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []WhiteWaterPoint{}, err
	}
	var points []WhiteWaterPoint
	err = json.Unmarshal(body, &points)
	if err != nil {
		return []WhiteWaterPoint{}, err
	}
	return points, nil
}

func SingleRouteTileHandler(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET, OPTIONS")

	callback := req.FormValue("callback")
	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	tracks := storage.FindTracksForRoute(id)
	points := storage.FindEventPointsForRoute(id)

	featureCollection := MkFeatureCollection(append(pointsToYmaps(points), tracksToYmaps(tracks)...))
	log.Infof("Found %d", len(featureCollection.Features))

	w.Write(JsonpAnswer(callback, featureCollection, "{}"))
}

func RouteEditorPageHandler(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET, OPTIONS")

	id, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}
	route := Route{}
	found, err := storage.FindRoute(id, &route)
	if found {
		tracks := storage.FindTracksForRoute(route.Id)
		points := storage.FindEventPointsForRoute(route.Id)
		bytes, err := json.Marshal(RouteEditorPage{
			Title:route.Title,
			Bounds: Bounds(tracks, points),
			Tracks: tracks,
			EventPoints:points,
			Category:route.Category,
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

	tracks := storage.FindTrackAsList(id)
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
		tilePoint, tileX, tileY := ToTileCoords(uint32(z), point)
		if x == tileX && y == tileY {
			features = append(features, Feature{
				Properties:FeatureProperties{
					HotspotMetaData : HotspotMetaData{
						Id:int64(i * 30 + int(z)),
						RenderedGeometry: NewYRectangleInt([][]int{
							{tilePoint.X - 15, tilePoint.Y - 15, },
							{tilePoint.X + 15, tilePoint.Y + 15, },
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

func GetVisibleRoutes(w http.ResponseWriter, req *http.Request) {
	corsHeaders(w, "GET")

	bbox, err := NewBbox(req.FormValue("bbox"))
	if err != nil {
		onError500(w, err, "Can not parse bbox")
		return
	}

	routes := storage.ListRoutes(bbox)
	log.Infof("Found %d", len(routes))

	routeInfos := make([]RouteEditorPage, len(routes))
	for i := 0; i < len(routes); i++ {
		route := routes[i]
		routeInfos[i] = RouteEditorPage{
			Id:route.Id,
			Title:route.Title,
			Tracks:storage.FindTracksForRoute(route.Id),
			EventPoints:storage.FindEventPointsForRoute(route.Id),
			Category: route.Category,
		}
	}

	w.Write([]byte(JsonStr(routeInfos, "[]")))
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

func CorsGetOptionsStub(w http.ResponseWriter, r *http.Request) {
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

func DelTrack(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = storage.DeleteTrack(id)
	if err != nil {
		onError500(w, err, "Can not remove track")
		return
	}
}

func EditRoute(w http.ResponseWriter, r *http.Request) {
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

	route, err := parseRouteForm(w, r)
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}
	route.Id = id

	err = storage.UpdateRoute(route)
	if err != nil {
		onError500(w, err, "Can not edit route	")
		return
	}
	writeRouteToResponse(id, w)
}

func DelRoute(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST, GET, OPTIONS, PUT, DELETE")

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		onError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = storage.DeleteTracksForRoute(id)
	if err != nil {
		onError500(w, err, "Can not remove tracks")
		return
	}
	err = storage.DeleteEventPointsForRoute(id)
	if err != nil {
		onError500(w, err, "Can not remove points")
		return
	}
	err = storage.DeleteRoute(id)
	if err != nil {
		onError500(w, err, "Can not remove route")
		return
	}
}

func writeRouteToResponse(id int64, w http.ResponseWriter) {
	route := Route{}
	found, err := storage.FindRoute(id, &route)
	if err != nil {
		onError500(w, err, "Can not find")
		return
	}
	if !found {
		onError(w, fmt.Errorf("Route with id %d does not exist", id), "Not found", http.StatusNotFound)
	}
	bytes, err := json.Marshal(route)
	if err != nil {
		onError500(w, err, "Can not marshal")
		return
	}
	w.Write(bytes)
}

func writeTrackToResponse(id int64, w http.ResponseWriter) {
	track := Track{}
	found, err := storage.FindTrack(id, &track)
	if err != nil {
		onError500(w, err, "Can not find")
		return
	}
	if !found {
		onError(w, fmt.Errorf("Track with id %d does not exist", id), "Not found", http.StatusNotFound)
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
	tType, err := ParseTrackType(r.FormValue("type"))
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

func parseRouteForm(w http.ResponseWriter, r *http.Request) (Route, error) {
	title := r.FormValue("title")
	category := model.SportCategory{}
	err := json.Unmarshal([]byte(r.FormValue("category")), &category)
	if err != nil {
		return Route{}, err
	}
	route := Route{
		Title:title,
		Category: category,
	}
	return route, nil
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

	eventPoint, routeId, err := parsePointForm(w, r)
	if err != nil {
		return
	}

	id, err := storage.AddEventPoint(routeId, eventPoint)

	if err != nil {
		onError500(w, err, "Can not insert")
		return
	}

	w.Write([]byte(strconv.FormatInt(id, 10)))
}

func parsePointForm(w http.ResponseWriter, r *http.Request) (EventPoint, int64, error) {
	route_id, err := strconv.ParseInt(r.FormValue("route_id"), 10, 64)
	if err != nil {
		onError(w, err, "Can not parse form", http.StatusBadRequest)
		return EventPoint{}, 0, err
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	pType, err := ParseEventPointType(r.FormValue("type"))
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

	return eventPoint, route_id, nil
}

func UploadTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		onError500(w, err, "Can not parse form")
		return
	}

	title := r.FormValue("title")
	redirectTo := r.FormValue("redirect_to")
	cookieDomain := r.FormValue("cookie_domain")
	cookiePath := r.FormValue("cookie_path")

	route := Route{
		Title: title,
	}
	routeId, err := storage.AddRoute(route)
	if err != nil {
		onError500(w, err, "Can not create route")
		return
	}

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
	tracks, points, err := parser.GetTracksAndPoints()
	if err != nil {
		onError500(w, err, "Bad geo data symantics")
		return
	}
	err = storage.AddTracks(routeId, tracks...)
	if err != nil {
		onError500(w, err, "Can not insert tracks")
		return
	}
	err = storage.AddEventPoints(routeId, points...)
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

func AddReport(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w, "POST")

	comment := r.FormValue("comment")
	objectIdStr := r.FormValue("object_id")
	objectId, err := strconv.ParseInt(objectIdStr, 10, 64)
	if err != nil {
		onError(w, err, fmt.Sprintf("Can not parse object id: %s", objectIdStr), 400)
		return
	}
	err = storage.AddReport(Report{
		ObjectId: objectId,
		Comment: comment,
	})
	if err != nil {
		onError500(w, err, "Can not add report")
		return
	}
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

	configuration := config.Load("")

	storage = NewPostgresStorage(configuration.DbConnString)
	clusterMaker = ClusterMaker{
		BarrierDistance: configuration.ClusterizationParams.BarrierRatio,
		MinDistance: configuration.ClusterizationParams.MinDistRatio,
	}

	r := mux.NewRouter()
	r.HandleFunc("/ymaps-tile", TileHandler)
	r.HandleFunc("/ymaps-single-route-tile", SingleRouteTileHandler)
	//r.HandleFunc("/track-active-areas/{id:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}/{z:[0-9]+}", TrackPointsToClickHandler)
	r.HandleFunc("/visible-routes", GetVisibleRoutes)

	r.HandleFunc("/route/{id}", CorsGetOptionsStub).Methods("GET", "OPTIONS")
	r.HandleFunc("/route/{id}", EditRoute).Methods("PUT")
	r.HandleFunc("/route/{id}", DelRoute).Methods("DELETE")
	r.HandleFunc("/route-editor-page", RouteEditorPageHandler)

	r.HandleFunc("/track/{id}", EditTrack).Methods("PUT")
	r.HandleFunc("/track/{id}", DelTrack).Methods("DELETE")
	r.HandleFunc("/track/{id}", CorsGetOptionsStub).Methods("GET", "OPTIONS")

	r.HandleFunc("/upload-track", UploadTrack).Methods("POST")

	r.HandleFunc("/point", AddPoint).Methods("POST")
	r.HandleFunc("/point/{id}", EditPoint).Methods("PUT")
	r.HandleFunc("/point/{id}", DelPoint).Methods("DELETE")
	r.HandleFunc("/point/{id}", GetPoint).Methods("OPTIONS", "GET")

	r.HandleFunc("/picture-metadata", PictureMetadataHandler).Methods("POST")

	r.HandleFunc("/ymaps-tile-ww", TileWhiteWaterHandler)
	r.HandleFunc("/whitewater", CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/whitewater", AddWhiteWaterPoints).Methods("PUT", "POST")
	r.HandleFunc("/nearest-rivers", GetNearestRivers).Methods("GET")
	r.HandleFunc("/visible-rivers", GetVisibleRivers).Methods("GET")

	r.HandleFunc("/gpx/{id}", DownloadGpx).Methods("GET")

	r.HandleFunc("/report", AddReport).Methods("POST")

	httpStr := fmt.Sprintf(":%d", 7007)
	log.Infof("Starting http server on %s", httpStr)
	http.Handle("/", r)
	err := http.ListenAndServe(httpStr, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
