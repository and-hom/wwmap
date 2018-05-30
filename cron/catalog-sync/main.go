package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"time"
)

func main() {
	log.Infof("Starting wwmap")
	configuration := config.Load("")
	pgStorage := dao.NewPostgresStorage(configuration.DbConnString)
	voyageReportDao := dao.VoyageReportStorage{pgStorage.(dao.PostgresStorage)}
	riverDao := dao.RiverStorage{pgStorage.(dao.PostgresStorage)}

	lastId, err := voyageReportDao.GetLastId()
	if err != nil {
		log.Fatalf("Can not connect get last report id: ", err.Error())
	}
	lastReportIdStr := lastId.(time.Time).Format(huskytm.TIME_FORMAT)
	log.Infof("Get and store reports since %s", lastReportIdStr)

	reportProvider, err := huskytm.GetReportProvider(configuration.Sync.Login, configuration.Sync.Password)
	if err != nil {
		log.Fatalf("Can not connect to source: ", err.Error())
	}
	defer reportProvider.Close()

	reports, next, err := reportProvider.ReportsSince(lastReportIdStr)
	if err != nil {
		log.Fatal("Can not get posts: ", err.Error())
	}
	if len(reports) == 0 {
		next = lastReportIdStr
	}

	reports, err = voyageReportDao.UpsertVoyageReports(reports...)
	if err != nil {
		log.Fatalf("Can not store reports: %v\n%s", reports, err.Error())
	}

	log.Infof("%d reports are successfully stored. Next id is %s\n", len(reports), next)

	log.Info("Now try to connect reports with known rivers")

	for _, report := range reports {
		log.Infof("Tags are: %v", report.Tags)
		rivers, err := riverDao.FindTitles(report.Tags)
		if err != nil {
			log.Fatal("Can not find rivers for tags", report.Tags, err)
		}
		log.Info(rivers)
		for _, river := range rivers {
			err := voyageReportDao.AssociateWithRiver(report.Id, river.Id)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}