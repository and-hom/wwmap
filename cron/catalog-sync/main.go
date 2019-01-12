package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm"
	"github.com/and-hom/wwmap/cron/catalog-sync/libru"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"github.com/and-hom/wwmap/cron/catalog-sync/tlib"
	"github.com/and-hom/wwmap/cron/catalog-sync/pdf"
	"github.com/and-hom/wwmap/lib/blob"
	"fmt"
	"flag"
	"strings"
	"github.com/and-hom/wwmap/lib/notification"
)

type App struct {
	CountryDao         dao.CountryDao
	RegionDao          dao.RegionDao
	RiverDao           dao.RiverDao
	WhiteWaterDao      dao.WhiteWaterDao
	ImgDao             dao.ImgDao
	VoyageReportDao    dao.VoyageReportDao
	NotificationDao    dao.NotificationDao
	WwPassportDao      dao.WwPassportDao
	Configuration      config.WordpressSync
	Notifications      config.Notifications

	ImgUrlBase         string
	ImgUrlPreviewBase  string
	ResourceBase       string

	NotificationHelper notification.NotificationHelper

	stat               *ImportExportReport
	reportProviders    []common.WithReportProvider
	catalogConnectors  []common.WithCatalogConnector

	sourceOnly         string
}

func CreateApp() App {
	configuration := config.Load("")
	configuration.ChangeLogLevel()

	pgStorage := dao.NewPostgresStorage(configuration.Db)
	riverPassportPdfStorage := blob.BasicFsStorage{
		BaseDir:configuration.RiverPassportPdfStorage.Dir,
	}
	riverPassportHtmlStorage := blob.BasicFsStorage{
		BaseDir:configuration.RiverPassportHtmlStorage.Dir,
	}
	userDao := dao.NewUserPostgresDao(pgStorage)
	notificationDao := dao.NewNotificationPostgresDao(pgStorage)
	return App{
		VoyageReportDao:dao.NewVoyageReportPostgresDao(pgStorage),
		CountryDao:dao.NewCountryPostgresDao(pgStorage),
		RegionDao:dao.NewRegionPostgresDao(pgStorage),
		RiverDao:dao.NewRiverPostgresDao(pgStorage),
		WhiteWaterDao:dao.NewWhiteWaterPostgresDao(pgStorage),
		ImgDao:dao.NewImgPostgresDao(pgStorage),
		WwPassportDao:dao.NewWWPassportPostgresDao(pgStorage),
		NotificationDao:notificationDao,
		Configuration:configuration.Sync,
		Notifications:configuration.Notifications,
		stat: &ImportExportReport{},
		reportProviders:[]common.WithReportProvider{
			common.WithReportProvider(func() (common.ReportProvider, error) {
				return huskytm.GetReportProvider(configuration.Sync.Login, configuration.Sync.Password, configuration.Sync.MinDeltaBetweenRequests)
			}),
			common.WithReportProvider(tlib.GetReportProvider),
			common.WithReportProvider(libru.GetReportProvider),
		},
		catalogConnectors: []common.WithCatalogConnector{
			{F:func() (common.CatalogConnector, error) {
				return huskytm.GetCatalogConnector(configuration.Sync.Login, configuration.Sync.Password, configuration.Sync.MinDeltaBetweenRequests)
			}},
			{F:func() (common.CatalogConnector, error) {
				return pdf.GetCatalogConnector(riverPassportPdfStorage, riverPassportHtmlStorage, configuration.RiverPassportPdfStorage.UrlBase)
			}},
		},
		ImgUrlBase:configuration.ImgStorage.Full.UrlBase,
		ImgUrlPreviewBase:configuration.ImgStorage.Preview.UrlBase,
		ResourceBase:configuration.Content.ResourceBase,
		NotificationHelper:notification.NotificationHelper{
			UserDao:userDao,
			NotificationDao:notificationDao,
			FallbackEmailRecipient:configuration.Notifications.FallbackEmailRecipient,
		},
	}
}

type Stage string

const SYNC_REPORTS Stage = "sync-reports"
const CATALOG_UP Stage = "catalog-up"
const ALL Stage = ""

func parseStage(s string) (Stage, error) {
	switch s {
	case string(SYNC_REPORTS):
		return SYNC_REPORTS, nil
	case string(CATALOG_UP):
		return CATALOG_UP, nil
	case "":
		return ALL, nil
	default:
		return Stage(""), fmt.Errorf("%s is not valid stage. Available are: sync-reports, catalog-up", s)
	}
}

func parseFlags(app App) (Stage, string, error) {
	stageStr := flag.String("stage", "", "Run only selected stage. Available are: sync-reports, catalog-up")

	sourceIdsMap := make(map[string]bool)
	sourceIds := []string{}
	for _, c := range app.catalogConnectors {
		sourceIdsMap[c.SourceId()] = true
	}
	for _, r := range app.reportProviders {
		sourceIdsMap[r.SourceId()] = true
	}
	for id, _ := range sourceIdsMap {
		sourceIds = append(sourceIds, id)
	}
	sourcePtr := flag.String("source", "", "Run only selected source. Available are: " + strings.Join(sourceIds, ", "))

	flag.Parse()
	log.Debug("stage=", *stageStr)
	log.Debug("source=", *sourcePtr)
	source := *sourcePtr

	stage, err := parseStage(*stageStr)
	if err != nil {
		return Stage(""), "", err
	}

	if source != "" {
		_, found := sourceIdsMap[source]
		if !found {
			return Stage(""), "", fmt.Errorf("Unknown source " + source + ". Available are: " + strings.Join(sourceIds, ","))
		}
	}

	return stage, source, err

}

func main() {
	log.Infof("Starting wwmap")
	app := CreateApp()

	stage, source, err := parseFlags(app);
	if err != nil {
		log.Fatal(err)
	}
	app.sourceOnly = source
	log.Infof("Stage=%s source=%s", string(stage), source)

	if stage == ALL || stage == SYNC_REPORTS {
		log.Info("Sync reports")
		app.DoSyncReports()
	}
	if stage == ALL || stage == CATALOG_UP {
		log.Info("Write catalog")
		app.DoWriteCatalog()
	}
}