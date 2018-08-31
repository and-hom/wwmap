package handler

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
	"fmt"
	"errors"
	"strconv"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
)

type TrackHandler struct {
	App
}

func (this *TrackHandler) UploadTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		OnError500(w, err, "Can not parse form")
		return
	}

	title := r.FormValue("title")
	redirectTo := r.FormValue("redirect_to")
	cookieDomain := r.FormValue("cookie_domain")
	cookiePath := r.FormValue("cookie_path")

	route := Route{
		Title: title,
	}
	routeId, err := this.Storage.AddRoute(route)
	if err != nil {
		OnError500(w, err, "Can not create route")
		return
	}

	file, _, err := r.FormFile("geo_file")
	if err != nil {
		OnError500(w, err, "Can not read form file")
		return
	}

	parser, err := this.geoParser(file)
	if err != nil {
		OnError500(w, err, "Can not parse geo data")
		return
	}
	tracks, points, err := parser.GetTracksAndPoints()
	if err != nil {
		OnError500(w, err, "Bad geo data symantics")
		return
	}
	err = this.Storage.AddTracks(routeId, tracks...)
	if err != nil {
		OnError500(w, err, "Can not insert tracks")
		return
	}
	err = this.Storage.AddEventPoints(routeId, points...)
	if err != nil {
		OnError500(w, err, "Can not insert tracks")
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

func (this *TrackHandler) TrackPointsToClickHandler(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, GET, OPTIONS)

	callback := req.FormValue("callback")
	pathParams := mux.Vars(req)

	x, err := strconv.Atoi(pathParams["x"])
	if err != nil {
		OnError(w, err, "Can not parse x", http.StatusBadRequest)
		return
	}

	y, err := strconv.Atoi(pathParams["y"])
	if err != nil {
		OnError(w, err, "Can not parse y", http.StatusBadRequest)
		return
	}

	z, err := strconv.ParseUint(pathParams["z"], 10, 32)
	if err != nil {
		OnError(w, err, "Can not parse z", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	tracks := this.Storage.FindTrackAsList(id)
	var t Track
	if len(tracks) > 0 {
		t = tracks[0]
	} else {
		OnError(w, errors.New("Not found"), fmt.Sprintf("Can not find track id = %d", id), http.StatusNotFound)
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

	w.Write(this.JsonpAnswer(callback, featureCollection, "{}"))
}

func (this *TrackHandler) EditTrack(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, POST, GET, OPTIONS, PUT, DELETE)
	err := r.ParseForm()
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	track, err := this.parseTrackForm(w, r)
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return
	}
	track.Id = id

	err = this.Storage.UpdateTrack(track)
	if err != nil {
		OnError500(w, err, "Can not edit track	")
		return
	}
	this.writeTrackToResponse(id, w)
}

func (this *TrackHandler) DelTrack(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, POST, GET, OPTIONS, PUT, DELETE)

	pathParams := mux.Vars(r)
	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = this.Storage.DeleteTrack(id)
	if err != nil {
		OnError500(w, err, "Can not remove track")
		return
	}
}

func (this *TrackHandler) writeTrackToResponse(id int64, w http.ResponseWriter) {
	track := Track{}
	found, err := this.Storage.FindTrack(id, &track)
	if err != nil {
		OnError500(w, err, "Can not find")
		return
	}
	if !found {
		OnError(w, fmt.Errorf("Track with id %d does not exist", id), "Not found", http.StatusNotFound)
	}
	bytes, err := json.Marshal(track)
	if err != nil {
		OnError500(w, err, "Can not marshal")
		return
	}
	w.Write(bytes)
}

func (this *TrackHandler) parseTrackForm(w http.ResponseWriter, r *http.Request) (Track, error) {
	title := r.FormValue("title")
	tType, err := ParseTrackType(r.FormValue("type"))
	if err != nil {
		OnError(w, err, "Can not parse form", http.StatusBadRequest)
		return Track{}, err
	}
	track := Track{
		Type:tType,
		Title:title,
	}

	return track, nil
}