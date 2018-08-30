package main

import (
	"github.com/gorilla/mux"
	"strconv"
	"net/http"
	"fmt"
	"github.com/ptrv/go-gpx"
	"regexp"
	"strings"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/model"
)

type GpxHandler struct{ Handler };

func (this *GpxHandler) DownloadGpx(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, "GET")
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	transliterate := req.FormValue("tr") != ""

	whitewaterPoints, err := this.whiteWaterDao.ListByRiver(id)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not read whitewater points for river %s", id))
		return
	}

	waypoints := make([]gpx.Wpt, len(whitewaterPoints))
	for i := 0; i < len(whitewaterPoints); i++ {
		whitewaterPoint := whitewaterPoints[i]
		waypoints[i] = gpx.Wpt{
			Lat: whitewaterPoint.Point.Lat,
			Lon: whitewaterPoint.Point.Lon,
			Cmt: whitewaterPoint.Comment,
		}
		categoryString := catStr(whitewaterPoint.Category, transliterate)

		titleString := whitewaterPoint.Title
		if transliterate {
			titleString = cyrillicToTranslit(whitewaterPoint.Title)
		}

		waypoints[i].Name = categoryString + titleString
	}
	if len(whitewaterPoints) == 0 {
		OnError(w, nil, fmt.Sprintf("No whitewater points found for river with id %d", id), http.StatusNotFound)
		return
	}
	gpxData := gpx.Gpx{
		Waypoints: waypoints,
		Creator: "wwmap",
	}
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.gpx\"", whitewaterPoints[0].RiverTitle))
	w.Header().Add("Content-Type", "application/gpx+xml")

	xmlBytes := gpxData.ToXML()
	w.Write(xmlBytes)
}

func catStr(category model.SportCategory, translit bool) string {
	if category.Category == -1 {
		if translit {
			return "(Stop!)"
		} else {
			return "(Непроход)"
		}
	}
	if category.Category == 0 {
		return ""
	}
	if category.Sub == "" {
		return "(" + category.Serialize() + ")"
	}
	return "(" + category.Serialize() + ")"
}

var translitCharMap = map[string]string{
	"a":"a",
	"б":"b",
	"в":"v",
	"г":"g",
	"д":"d",
	"е":"e",
	"ё":"e",
	"ж":"zh",
	"з":"z",
	"и":"i",
	"й":"j",
	"к":"k",
	"л":"l",
	"м":"m",
	"н":"n",
	"о":"o",
	"п":"p",
	"р":"r",
	"с":"s",
	"т":"t",
	"у":"u",
	"ф":"f",
	"х":"h",
	"ц":"ts",
	"ч":"ch",
	"ш":"sh",
	"щ":"sch",
	"ы":"y",
	"ь":"'",
	"э":"ye",
	"ю":"ju",
	"я":"ya",
}

func doReplace(data string, from string, to string) string {
	r, _ := regexp.Compile(from)
	return r.ReplaceAllString(data, to)
}

func cyrillicToTranslit(cyrillicString string) string {
	translitString := cyrillicString
	for k, v := range translitCharMap {
		translitString = doReplace(translitString, k, v)
		translitString = doReplace(translitString, strings.ToUpper(k), strings.Title(v))
	}
	return translitString
}

