package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	. "github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/gorilla/handlers"
	"os"
	"github.com/and-hom/wwmap/backend/passport"
	"time"
)

type App struct {
	storage         Storage
	riverDao        RiverDao
	whiteWaterDao   WhiteWaterDao
	reportDao       ReportDao
	voyageReportDao VoyageReportDao
	imgDao          ImgDao
	userDao         UserDao
	countryDao      CountryDao
	regionDao       RegionDao
	yandexPassport  passport.YandexPassport
}

func main() {
	log.Infof("Starting wwmap")

	configuration := config.Load("")

	storage := NewPostgresStorage(configuration.DbConnString)

	riverDao := NewRiverPostgresDao(storage)
	voyageReportDao := NewVoyageReportPostgresDao(storage)
	imgDao := NewImgPostgresDao(storage)
	whiteWaterDao := NewWhiteWaterPostgresDao(storage)
	reportDao := NewReportPostgresDao(storage)
	userDao := NewUserPostgresDao(storage)
	countryDao := NewCountryPostgresDao(storage)
	regionDao := NewRegionPostgresDao(storage)

	yandexPassport := passport.New(15 * time.Minute)

	clusterMaker := NewClusterMaker(whiteWaterDao, imgDao,
		configuration.ClusterizationParams)

	app := App{
		storage:&storage,
		riverDao:riverDao,
		whiteWaterDao:whiteWaterDao,
		reportDao:reportDao,
		voyageReportDao: voyageReportDao,
		imgDao:imgDao,
		userDao: userDao,
		countryDao: countryDao,
		regionDao: regionDao,
		yandexPassport: yandexPassport,
	}

	handler := Handler{app}
	gpxHandler := GpxHandler{handler}
	routeHandler := RouteHandler{handler}
	trackHandler := TrackHandler{handler}
	riverHandler := RiverHandler{handler, configuration.Content.ResourceBase}
	whiteWaterHandler := WhiteWaterHandler{Handler:handler, resourceBase:configuration.Content.ResourceBase, clusterMaker: clusterMaker}
	pointHandler := PointHandler{handler}
	reportHandler := ReportHandler{handler}
	pictureHandler := PictureHandler{handler}
	userInfoHandler := UserInfoHandler{handler}
	geoHierarchyHandler := GeoHierarchyHandler{handler}

	r := mux.NewRouter()
	r.HandleFunc("/ymaps-tile", handler.TileHandler)
	r.HandleFunc("/ymaps-single-route-tile", routeHandler.SingleRouteTileHandler)
	//r.HandleFunc("/track-active-areas/{id:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}/{z:[0-9]+}", handler.TrackPointsToClickHandler)
	r.HandleFunc("/visible-routes", routeHandler.GetVisibleRoutes)

	r.HandleFunc("/route/{id}", handler.CorsGetOptionsStub).Methods("GET", "OPTIONS")
	r.HandleFunc("/route/{id}", routeHandler.EditRoute).Methods("PUT")
	r.HandleFunc("/route/{id}", routeHandler.DelRoute).Methods("DELETE")
	r.HandleFunc("/route-editor-page", routeHandler.RouteEditorPageHandler)

	r.HandleFunc("/track/{id}", trackHandler.EditTrack).Methods("PUT")
	r.HandleFunc("/track/{id}", trackHandler.DelTrack).Methods("DELETE")
	r.HandleFunc("/track/{id}", handler.CorsGetOptionsStub).Methods("GET", "OPTIONS")

	r.HandleFunc("/upload-track", trackHandler.UploadTrack).Methods("POST")

	r.HandleFunc("/point", pointHandler.AddPoint).Methods("POST")
	r.HandleFunc("/point/{id}", pointHandler.EditPoint).Methods("PUT")
	r.HandleFunc("/point/{id}", pointHandler.DelPoint).Methods("DELETE")
	r.HandleFunc("/point/{id}", pointHandler.GetPoint).Methods("OPTIONS", "GET")

	r.HandleFunc("/picture-metadata", pictureHandler.PictureMetadataHandler).Methods("POST")

	r.HandleFunc("/ymaps-tile-ww", whiteWaterHandler.TileWhiteWaterHandler)
	r.HandleFunc("/whitewater", whiteWaterHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/whitewater", whiteWaterHandler.AddWhiteWaterPoints).Methods("PUT", "POST")

	r.HandleFunc("/nearest-rivers", riverHandler.GetNearestRivers).Methods("GET")
	r.HandleFunc("/visible-rivers", riverHandler.GetVisibleRivers).Methods("GET")

	r.HandleFunc("/gpx/{id}", gpxHandler.DownloadGpx).Methods("GET")

	r.HandleFunc("/report", reportHandler.AddReport).Methods("POST")

	r.HandleFunc("/user-info", userInfoHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/user-info", userInfoHandler.GetUserInfo).Methods("GET")
	r.HandleFunc("/auth-test", userInfoHandler.TestAuth).Methods("GET")

	// Web editor support
	r.HandleFunc("/country", geoHierarchyHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/country", geoHierarchyHandler.ListCountries).Methods("GET")
	r.HandleFunc("/country/{countryId}/region", geoHierarchyHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/country/{countryId}/region", geoHierarchyHandler.ListRegions).Methods("GET")
	r.HandleFunc("/country/{countryId}/region/{regionId}/river", geoHierarchyHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/country/{countryId}/region/{regionId}/river", geoHierarchyHandler.ListRegionRivers).Methods("GET")
	r.HandleFunc("/country/{countryId}/river", geoHierarchyHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/country/{countryId}/river", geoHierarchyHandler.ListCountryRivers).Methods("GET")
	r.HandleFunc("/river/{riverId}", geoHierarchyHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/river/{riverId}", geoHierarchyHandler.GetRiver).Methods("GET")
	r.HandleFunc("/river/{riverId}/reports", geoHierarchyHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/river/{riverId}/reports", geoHierarchyHandler.ListRiverReports).Methods("GET")
	r.HandleFunc("/river/{riverId}/spots", geoHierarchyHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/river/{riverId}/spots", geoHierarchyHandler.ListSpots).Methods("GET")
	r.HandleFunc("/region", geoHierarchyHandler.CorsGetOptionsStub).Methods("OPTIONS")
	r.HandleFunc("/region", geoHierarchyHandler.ListAllRegions).Methods("GET")

	log.Infof("Starting http server on %s", configuration.Api.BindTo)
	http.Handle("/", r)
	err := http.ListenAndServe(configuration.Api.BindTo, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
