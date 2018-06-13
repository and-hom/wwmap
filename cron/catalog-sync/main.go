package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
)

type App struct {
	VoyageReportDao  dao.VoyageReportDao
	WhiteWaterDao    dao.WhiteWaterDao
	RiverDao         dao.RiverDao
	ImgDao           dao.ImgDao
	WwPassportDao    dao.WwPassportDao
	Configuration    config.WordpressSync
	Notifications    config.Notifications

	stat             *ImportExportReport
	catalogConnector common.CatalogConnector
}

func CreateApp() App {
	configuration := config.Load("")
	pgStorage := dao.NewPostgresStorage(configuration.DbConnString)
	return App{
		VoyageReportDao:dao.VoyageReportStorage{pgStorage.(dao.PostgresStorage)},
		RiverDao:dao.RiverStorage{pgStorage.(dao.PostgresStorage)},
		WhiteWaterDao:dao.WhiteWaterStorage{pgStorage.(dao.PostgresStorage)},
		ImgDao:dao.ImgStorage{pgStorage.(dao.PostgresStorage)},
		WwPassportDao:dao.WWPassportStorage{pgStorage.(dao.PostgresStorage)},
		Configuration:configuration.Sync,
		Notifications:configuration.Notifications,
		stat: &ImportExportReport{},
	}
}

func main() {
	log.SetLevel(log.DebugLevel)
	log.Infof("Starting wwmap")
	app := CreateApp()
	app.DoSyncReports()
	//app.DoReadCatalog()
	//app.DoWriteCatalog()
}

func (this App) getCachedCatalogConnector() common.CatalogConnector {
	if this.catalogConnector == nil {
		catalogConnector, err := huskytm.GetCatalogConnector(this.Configuration.Login, this.Configuration.Password)
		if err != nil {
			this.Fatalf(err, "Can not connect to catalog")
		}
		this.catalogConnector = catalogConnector
	}
	return this.catalogConnector
}
