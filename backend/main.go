package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/gorilla/handlers"
	"os"
	"github.com/and-hom/wwmap/backend/passport"
	"time"
	"github.com/and-hom/wwmap/backend/referer"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/backend/handler"
	"github.com/and-hom/wwmap/backend/clustering"
)

func main() {
	log.Infof("Starting wwmap")

	configuration := config.Load("")
	configuration.ChangeLogLevel()

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

	clusterMaker := clustering.NewClusterMaker(whiteWaterDao, imgDao,
		configuration.ClusterizationParams)

	imgStorage := blob.BasicFsStorage{
		BaseDir:configuration.ImgStorage.Full.Dir,
	}
	imgPreviewStorage := blob.BasicFsStorage{
		BaseDir:configuration.ImgStorage.Preview.Dir,
	}
	riverPassportStorage := blob.BasicFsStorage{
		BaseDir:configuration.RiverPassportStorage.Dir,
	}

	app := handler.App{
		Handler: Handler{},
		Storage:&storage,
		RiverDao:riverDao,
		WhiteWaterDao:whiteWaterDao,
		ReportDao:reportDao,
		VoyageReportDao: voyageReportDao,
		ImgDao:imgDao,
		UserDao: userDao,
		CountryDao: countryDao,
		RegionDao: regionDao,
		YandexPassport: yandexPassport,
		RefererStorage: referer.CreateDbReferrerStorage(storage),
		ImgUrlBase:configuration.ImgStorage.Full.UrlBase,
		ImgUrlPreviewBase:configuration.ImgStorage.Preview.UrlBase,
	}

	_handlers := []ApiHandler{
		&handler.GpxHandler{app},
		&handler.RiverHandler{
			App:app,
			ResourceBase: configuration.Content.ResourceBase,
			RiverPassportUrlBase: configuration.RiverPassportStorage.UrlBase,
		},
		&handler.WhiteWaterHandler{App:app, ResourceBase:configuration.Content.ResourceBase, ClusterMaker: clusterMaker},
		&handler.ReportHandler{app},
		&handler.UserInfoHandler{app},
		&handler.GeoHierarchyHandler{
			App: app,
			ImgStorage: imgStorage,
			PreviewImgStorage: imgPreviewStorage,
			RiverPassportStorage:riverPassportStorage,
		},
		&handler.ImgHandler{
			App:app,
			ImgStorage: imgStorage,
			PreviewImgStorage: imgPreviewStorage,
		},
		&handler.RefSitesHandler{app},
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
