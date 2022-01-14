package main

import (
	"flag"
	"fmt"
	"github.com/and-hom/wwmap/cron/report-retriever/common"
	"github.com/and-hom/wwmap/cron/report-retriever/huskytm"
	"github.com/and-hom/wwmap/cron/report-retriever/libru"
	"github.com/and-hom/wwmap/cron/report-retriever/riskru"
	"github.com/and-hom/wwmap/cron/report-retriever/skitalets"
	"github.com/and-hom/wwmap/cron/report-retriever/tlib"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/notification"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type App struct {
	RiverDao        dao.RiverDao
	WhiteWaterDao   dao.WhiteWaterDao
	ImgDao          dao.ImgDao
	VoyageReportDao dao.VoyageReportDao
	LevelSensorDao  dao.LevelSensorDao
	LevelDao        dao.LevelDao

	NotificationHelper notification.NotificationHelper

	stat            *ImportExportReport
	reportProviders []common.WithReportProvider

	sourceOnly string
	rateLimit  util.RateLimit
}

func CreateApp() App {
	configuration := config.Load("")
	configuration.ConfigureLogger()

	pgStorage := dao.NewPostgresStorage(configuration.Db)
	userDao := dao.NewUserPostgresDao(pgStorage)
	notificationDao := dao.NewNotificationPostgresDao(pgStorage)
	return App{
		VoyageReportDao: dao.NewVoyageReportPostgresDao(pgStorage),
		RiverDao:        dao.NewRiverPostgresDao(pgStorage),
		WhiteWaterDao:   dao.NewWhiteWaterPostgresDao(pgStorage),
		ImgDao:          dao.NewImgPostgresDao(pgStorage),
		LevelDao:        dao.NewLevelPostgresDao(pgStorage),
		LevelSensorDao:  dao.NewLevelSensorPostgresDao(pgStorage),
		stat:            &ImportExportReport{},
		reportProviders: []common.WithReportProvider{
			func() (common.ReportProvider, error) {
				return huskytm.GetReportProvider(
					configuration.Sync.Login,
					configuration.Sync.Password,
					configuration.Sync.MinDeltaBetweenRequests,
				)
			},
			tlib.GetReportProvider,
			libru.GetReportProvider,
			skitalets.GetReportProvider,
			riskru.GetReportProvider,
		},
		NotificationHelper: notification.NotificationHelper{
			UserDao:                userDao,
			NotificationDao:        notificationDao,
			FallbackEmailRecipient: configuration.Notifications.FallbackEmailRecipient,
		},
		rateLimit: util.NewRateLimit(time.Second),
	}
}

func parseFlags(app App) (string, error) {
	sourceIdsMap := make(map[string]bool)
	sourceIds := []string{}
	for _, r := range app.reportProviders {
		sourceIdsMap[r.SourceId()] = true
	}
	for id, _ := range sourceIdsMap {
		sourceIds = append(sourceIds, id)
	}
	sourcePtr := flag.String("source", "", "Run only selected source. Available are: "+strings.Join(sourceIds, ", "))

	flag.Parse()
	log.Debug("source=", *sourcePtr)
	source := *sourcePtr

	if source != "" {
		_, found := sourceIdsMap[source]
		if !found {
			return "", fmt.Errorf("Unknown source " + source + ". Available are: " + strings.Join(sourceIds, ","))
		}
	}

	return source, nil

}

func main() {
	exif.RegisterParsers(mknote.All...)

	log.Infof("Starting wwmap")
	app := CreateApp()

	source, err := parseFlags(app)
	if err != nil {
		log.Fatal(err)
	}
	app.sourceOnly = source

	if source == "" {
		log.Info("Sync reports for all sources")
	} else {
		log.Infof("Sync reports for %s", source)
	}
	app.DoSyncReports()

}
