package handler

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	. "github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/and-hom/wwmap/backend/referer"
	"github.com/and-hom/wwmap/backend/ymaps"
)

type App struct {
	Handler
	Storage         Storage
	RiverDao        RiverDao
	WhiteWaterDao   WhiteWaterDao
	ReportDao       ReportDao
	VoyageReportDao VoyageReportDao
	ImgDao          ImgDao
	UserDao         UserDao
	CountryDao      CountryDao
	RegionDao       RegionDao
	YandexPassport  passport.YandexPassport
	RefererStorage  referer.RefererStorage
}


func (this *App) TileHandler(w http.ResponseWriter, req *http.Request) {
	callback, bbox, err := this.tileParams(w, req)
	if err != nil {
		return
	}
	tracks := this.Storage.ListTracks(bbox)
	points := this.Storage.ListPoints(bbox)

	featureCollection := MkFeatureCollection(append(ymaps.PointsToYmaps(points), ymaps.TracksToYmaps(tracks)...))
	log.Infof("Found %d", len(featureCollection.Features))

	w.Write(this.JsonpAnswer(callback, featureCollection, "{}"))
}

