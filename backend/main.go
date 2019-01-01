package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/backend/passport"
	"time"
	"github.com/and-hom/wwmap/backend/referer"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/backend/handler"
	"github.com/and-hom/wwmap/backend/clustering"
	"github.com/and-hom/wwmap/lib/notification"
	"os"
	"github.com/gorilla/handlers"
)

var version string = "development"

func main() {
	log.Infof("Starting wwmap")

	configuration := config.Load("")
	configuration.ChangeLogLevel()

	storage := NewPostgresStorage(configuration.Db)

	riverDao := NewRiverPostgresDao(storage)
	voyageReportDao := NewVoyageReportPostgresDao(storage)
	imgDao := NewImgPostgresDao(storage)
	whiteWaterDao := NewWhiteWaterPostgresDao(storage)
	notificationDao := NewNotificationPostgresDao(storage)
	userDao := NewUserPostgresDao(storage)
	countryDao := NewCountryPostgresDao(storage)
	regionDao := NewRegionPostgresDao(storage)
	tileDao := NewTilePostgresDao(storage)

	clusterMaker := clustering.NewClusterMaker(configuration.ClusterizationParams)

	imgStorage := blob.BasicFsStorage{
		BaseDir:configuration.ImgStorage.Full.Dir,
	}
	imgPreviewStorage := blob.BasicFsStorage{
		BaseDir:configuration.ImgStorage.Preview.Dir,
	}
	riverPassportPdfStorage := blob.BasicFsStorage{
		BaseDir:configuration.RiverPassportPdfStorage.Dir,
	}
	riverPassportHtmlStorage := blob.BasicFsStorage{
		BaseDir:configuration.RiverPassportHtmlStorage.Dir,
	}

	r := mux.NewRouter()

	app := handler.App{
		Handler: Handler{R:r},
		Storage:&storage,
		RiverDao:riverDao,
		WhiteWaterDao:whiteWaterDao,
		NotificationDao:notificationDao,
		VoyageReportDao: voyageReportDao,
		ImgDao:imgDao,
		UserDao: userDao,
		CountryDao: countryDao,
		RegionDao: regionDao,
		TileDao:tileDao,
		AuthProviders: map[AuthProvider]passport.Passport{
			YANDEX: passport.Yandex(15 * time.Minute),
			GOOGLE:     passport.Google(15 * time.Minute),
			VK:     passport.Vk(15 * time.Minute),
		},
		RefererStorage: referer.CreateDbReferrerStorage(storage),
		ImgUrlBase:configuration.ImgStorage.Full.UrlBase,
		ImgUrlPreviewBase:configuration.ImgStorage.Preview.UrlBase,
		NotificationHelper:notification.NotificationHelper{
			NotificationDao:notificationDao,
			UserDao: userDao,
			FallbackEmailRecipient:configuration.Notifications.FallbackEmailRecipient,
		},
	}

	_handlers := []ApiHandler{
		&handler.GpxHandler{app},
		&handler.RiverHandler{
			App:app,
			ResourceBase: configuration.Content.ResourceBase,
			RiverPassportPdfUrlBase: configuration.RiverPassportPdfStorage.UrlBase,
			RiverPassportHtmlUrlBase: configuration.RiverPassportHtmlStorage.UrlBase,
		},
		&handler.WhiteWaterHandler{App:app, ResourceBase:configuration.Content.ResourceBase, ClusterMaker: clusterMaker},
		&handler.ReportHandler{app},
		&handler.UserInfoHandler{app},
		&handler.GeoHierarchyHandler{
			App: app,
			ImgStorage: imgStorage,
			PreviewImgStorage: imgPreviewStorage,
			RiverPassportHtmlStorage:riverPassportHtmlStorage,
			RiverPassportPdfStorage:riverPassportPdfStorage,
		},
		&handler.ImgHandler{
			App:app,
			ImgStorage: imgStorage,
			PreviewImgStorage: imgPreviewStorage,
		},
		&handler.RefSitesHandler{app},
		handler.CreateSystemHandler(&app, version),
	}

	for _, h := range _handlers {
		h.Init()
	}

	log.Infof("Starting http server on %s", configuration.Api.BindTo)
	http.Handle("/", r)

	err := http.ListenAndServe(configuration.Api.BindTo, createHttpHandler(configuration))
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}

func createHttpHandler(configuration config.Configuration) http.Handler {
	var h http.Handler = http.DefaultServeMux
	logLevel, err := configuration.LogLevel.ToLogrus()
	if err != nil {
		log.Fatalf("Can not parse log level %s", configuration.LogLevel)
	}
	if logLevel == log.DebugLevel {
		h = handlers.LoggingHandler(os.Stdout, h)
	}
	return h
}
