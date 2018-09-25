package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"github.com/and-hom/wwmap/cron/catalog-sync/tlib"
	"github.com/and-hom/wwmap/cron/catalog-sync/pdf"
)

type App struct {
	CountryDao        dao.CountryDao
	RegionDao         dao.RegionDao
	RiverDao          dao.RiverDao
	WhiteWaterDao     dao.WhiteWaterDao
	ImgDao            dao.ImgDao
	VoyageReportDao   dao.VoyageReportDao
	WwPassportDao     dao.WwPassportDao
	Configuration     config.WordpressSync
	Notifications     config.Notifications

	ImgUrlBase        string
	ImgUrlPreviewBase string
	ResourceBase      string

	stat              *ImportExportReport
	catalogConnector  common.CatalogConnector
	reportProviders   []common.WithReportProvider
	catalogConnectors []common.WithCatalogConnector
}

func CreateApp() App {
	configuration := config.Load("")
	configuration.ChangeLogLevel()

	pgStorage := dao.NewPostgresStorage(configuration.DbConnString)
	return App{
		VoyageReportDao:dao.NewVoyageReportPostgresDao(pgStorage),
		CountryDao:dao.NewCountryPostgresDao(pgStorage),
		RegionDao:dao.NewRegionPostgresDao(pgStorage),
		RiverDao:dao.NewRiverPostgresDao(pgStorage),
		WhiteWaterDao:dao.NewWhiteWaterPostgresDao(pgStorage),
		ImgDao:dao.NewImgPostgresDao(pgStorage),
		WwPassportDao:dao.NewWWPassportPostgresDao(pgStorage),
		Configuration:configuration.Sync,
		Notifications:configuration.Notifications,
		stat: &ImportExportReport{},
		reportProviders:[]common.WithReportProvider{
			common.WithReportProvider(func() (common.ReportProvider, error) {
				return huskytm.GetReportProvider(configuration.Sync.Login, configuration.Sync.Password, configuration.Sync.MinDeltaBetweenRequests)
			}),
			common.WithReportProvider(tlib.GetReportProvider),
		},
		catalogConnectors: []common.WithCatalogConnector{
			//{F:func() (common.CatalogConnector, error) {
			//	return huskytm.GetCatalogConnector(configuration.Sync.Login, configuration.Sync.Password, configuration.Sync.MinDeltaBetweenRequests)
			//}},
			{F:pdf.GetCatalogConnector},
		},
		ImgUrlBase:configuration.ImgStorage.Full.UrlBase,
		ImgUrlPreviewBase:configuration.ImgStorage.Preview.UrlBase,
		ResourceBase:configuration.Content.ResourceBase,
	}
}

func main() {
	log.Infof("Starting wwmap")
	app := CreateApp()
	//app.DoSyncReports()
	app.DoWriteCatalog()
}