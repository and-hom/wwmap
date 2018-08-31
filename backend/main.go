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
	"github.com/and-hom/wwmap/backend/referer"
	"github.com/and-hom/wwmap/lib/img_storage"
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
	refererStorage  referer.RefererStorage
}

const (
	OPTIONS = "OPTIONS"
	HEAD = "HEAD"
	GET = "GET"
	PUT = "PUT"
	POST = "POST"
	DELETE = "DELETE"
)

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
		refererStorage: referer.CreateDbReferrerStorage(storage),
	}

	handler := Handler{app}
	_handlers := []ApiHandler{
		&GpxHandler{handler},
		&RiverHandler{handler, configuration.Content.ResourceBase},
		&WhiteWaterHandler{Handler:handler, resourceBase:configuration.Content.ResourceBase, clusterMaker: clusterMaker},
		&ReportHandler{handler},
		&UserInfoHandler{handler},
		&GeoHierarchyHandler{Handler: handler},
		&ImgHandler{
			Handler:handler,
			imgStorage: img_storage.BasicFsStorage{
				BaseDir:configuration.ImgStorage.Full.Dir,
			},
			previewImgStorage: img_storage.BasicFsStorage{
				BaseDir:configuration.ImgStorage.Preview.Dir,
			},
			imgUrlBase:configuration.ImgStorage.Full.UrlBase,
			imgUrlPreviewBase:configuration.ImgStorage.Preview.UrlBase,
		},
		&RefSitesHandler{handler},
	}

	r := mux.NewRouter()
	for _, h := range _handlers {
		h.Init(r)
	}

	log.Infof("Starting http server on %s", configuration.Api.BindTo)
	http.Handle("/", r)
	err := http.ListenAndServe(configuration.Api.BindTo, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
