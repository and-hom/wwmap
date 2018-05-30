package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"time"
	"strings"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
)

type App struct {
	VoyageReportDao dao.VoyageReportDao
	WhiteWaterDao   dao.WhiteWaterDao
	RiverDao        dao.RiverDao
	ImgDao          dao.ImgDao
	Configuration   config.WordpressSync
}

func CreateApp() App {
	configuration := config.Load("")
	pgStorage := dao.NewPostgresStorage(configuration.DbConnString)
	return App{
		VoyageReportDao:dao.VoyageReportStorage{pgStorage.(dao.PostgresStorage)},
		RiverDao:dao.RiverStorage{pgStorage.(dao.PostgresStorage)},
		WhiteWaterDao:dao.WhiteWaterStorage{pgStorage.(dao.PostgresStorage)},
		ImgDao:dao.ImgStorage{pgStorage.(dao.PostgresStorage)},
		Configuration:configuration.Sync,
	}
}

func main() {
	log.Infof("Starting wwmap")
	app := CreateApp()
	app.DoSync()
}

func (this App) DoSync() {
	lastId, err := this.VoyageReportDao.GetLastId()
	if err != nil {
		log.Fatalf("Can not connect get last report id: ", err.Error())
	}
	lastReportIdStr := lastId.(time.Time).Format(huskytm.TIME_FORMAT)
	log.Infof("Get and store reports since %s", lastReportIdStr)

	reportProvider, err := huskytm.GetReportProvider(this.Configuration.Login, this.Configuration.Password)
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

	reports, err = this.VoyageReportDao.UpsertVoyageReports(reports...)
	if err != nil {
		log.Fatalf("Can not store reports: %v\n%s", reports, err.Error())
	}

	log.Infof("%d reports are successfully stored. Next id is %s\n", len(reports), next)

	log.Info("Try to connect reports with known rivers")

	for _, report := range reports {
		log.Infof("Tags are: %v", report.Tags)
		rivers, err := this.RiverDao.FindTitles(report.Tags)
		if err != nil {
			log.Fatal("Can not find rivers for tags", report.Tags, err)
		}
		log.Info(rivers)
		for _, river := range rivers {
			err := this.VoyageReportDao.AssociateWithRiver(report.Id, river.Id)
			if err != nil {
				log.Fatal(err)
			}
		}
		report.Rivers = rivers
	}

	for _, report := range reports {
		this.findMatchAndStoreImages(report, reportProvider)
	}
}

func (this App) findMatchAndStoreImages(report dao.VoyageReport, reportProvider common.ReportProvider) {
	log.Infof("Find images for report %d", report.Id)
	imgs, err := reportProvider.Images(report.RemoteId)
	if err != nil {
		log.Fatalf("Can not load images for report %d: %s", report.Id, err.Error())
	}

	log.Infof("Bind images to ww spots for report %d", report.Id)
	matchedImgs := []dao.Img{}
	candidates := this.matchImgsToWhiteWaterPoints(report, imgs)

	for _, imgToWwpts := range candidates {
		if len(imgToWwpts.Wwpts) == 1 {
			imgToWwpts.Img.WwId = imgToWwpts.Wwpts[0].Id
			matchedImgs = append(matchedImgs, imgToWwpts.Img)
		} else if len(imgToWwpts.Wwpts) > 1 {
			log.Warn("More then one white water point for img ", imgToWwpts.Img.RemoteId)
		}
	}

	log.Infof("Store images for report %d", report.Id)
	_, err = this.ImgDao.Upsert(matchedImgs...)
	if err != nil {
		log.Fatalf("Can not upsert images for report %d: %s", report.Id, err.Error())
	}
}

type ImgWwPoints struct {
	Img   dao.Img
	Wwpts []dao.WhiteWaterPointWithRiverTitle
}

func (this App) matchImgsToWhiteWaterPoints(report dao.VoyageReport, imgs []dao.Img) map[string]ImgWwPoints {
	candidates := make(map[string]ImgWwPoints)
	for _, img := range imgs {
		for _, river := range report.Rivers {
			wwpts, err := this.WhiteWaterDao.ListWhiteWaterPointsByRiver(river.Id)
			if err != nil {
				log.Fatalf("Can not list white water spots for river %d: %s", river.Id, err.Error())
			}
			for _, wwpt := range wwpts {
				for _, label := range img.LabelsForSearch {
					if strings.Contains(strings.ToLower(label), strings.ToLower(wwpt.Title)) {
						rec, found := candidates[img.RemoteId]
						if !found {
							rec = ImgWwPoints{
								Img:img,
								Wwpts:[]dao.WhiteWaterPointWithRiverTitle{},
							}
							candidates[img.RemoteId] = rec
						}
						rec.Wwpts = append(rec.Wwpts, wwpt)
					}
				}
			}
		}
	}
	return candidates
}