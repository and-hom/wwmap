package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"github.com/gorilla/handlers"
	. "github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/config"
)

type App struct {
	storage       Storage
	riverDao      RiverDao
	whiteWaterDao WhiteWaterDao
	reportDao     ReportDao
	clusterMaker  ClusterMaker
}

func corsHeaders(w http.ResponseWriter, methods string) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", methods)
	w.Header().Add("Access-Control-Allow-Headers", "origin, x-csrftoken, content-type, accept")
}

func main() {
	log.Infof("Starting wwmap")

	configuration := config.Load("")

	storage := NewPostgresStorage(configuration.DbConnString)

	riverDao := RiverStorage{storage.(PostgresStorage)}
	whiteWaterDao := WhiteWaterStorage{storage.(PostgresStorage)}
	reportDao := ReportStorage{storage.(PostgresStorage)}

	clusterMaker := ClusterMaker{
		BarrierDistance: configuration.ClusterizationParams.BarrierRatio,
		MinDistance: configuration.ClusterizationParams.MinDistRatio,
		SinglePointClusteringMaxZoom: configuration.ClusterizationParams.SinglePointClusteringMaxZoom,
	}

	app := App{
		storage:storage,
		riverDao:riverDao,
		whiteWaterDao:whiteWaterDao,
		reportDao:reportDao,
		clusterMaker:clusterMaker,
	}

	handler := Handler{app}
	gpxHandler := GpxHandler{handler}
	routeHandler := RouteHandler{handler}
	trackHandler := TrackHandler{handler}
	riverHandler := RiverHandler{handler}
	whiteWaterHandler := WhiteWaterHandler{handler, configuration.Content.ResourceBase}
	pointHandler := PointHandler{handler}
	reportHandler := ReportHandler{handler}
	pictureHandler := PictureHandler{handler}

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

	log.Infof("Starting http server on %s", configuration.Api.BindTo)
	http.Handle("/", r)
	err := http.ListenAndServe(configuration.Api.BindTo, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
