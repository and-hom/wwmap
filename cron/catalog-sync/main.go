package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"time"
	"strings"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"io"
	"html/template"
	"github.com/and-hom/wwmap/lib/mail"
	"fmt"
)

type App struct {
	VoyageReportDao dao.VoyageReportDao
	WhiteWaterDao   dao.WhiteWaterDao
	RiverDao        dao.RiverDao
	ImgDao          dao.ImgDao
	WwPassportDao   dao.WwPassportDao
	Configuration   config.WordpressSync
	Notifications   config.Notifications
	stat            *ImportExportReport
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
	log.Infof("Starting wwmap")
	app := CreateApp()
	app.DoSyncReports()
	//app.DoSyncCatalog()
}

func (this App) DoSyncCatalog() {
	lastId, err := this.WwPassportDao.GetLastId(huskytm.SOURCE)
	if err != nil {
		this.Fatalf(err, "Can not connect get last ww passport entry id")
	}
	lastWwPassportIdStr := lastId.(time.Time).Format(huskytm.TIME_FORMAT)
	log.Infof("Get and store ww passport entries since %s", lastWwPassportIdStr)

	catalogConnector, err := huskytm.GetCatalogConnector(this.Configuration.Login, this.Configuration.Password)
	if err != nil {
		this.Fatalf(err, "Can not connect to catalog")
	}
	defer catalogConnector.Close()

	wwPassportEntries, err := catalogConnector.PassportEntriesSince(lastWwPassportIdStr)
	if err != nil {
		log.Fatal("Can not get posts: ", err.Error())
	}

	log.Info("")
	for _, wwPassport := range wwPassportEntries {
		rivers, err := this.RiverDao.FindTitles([]string{wwPassport.River})
		if err != nil {
			log.Fatal("Can not find rivers for ww passport", wwPassport.Source, wwPassport.RemoteId, wwPassport.River)
		}
		switch len(rivers) {
		case 1:
			wwpts, err := this.WhiteWaterDao.ListWhiteWaterPointsByRiverAndTitle(rivers[0].Id, wwPassport.Title)
			if err != nil {
				log.Fatal("Can not find ww points for ww passport", wwPassport.Source, wwPassport.RemoteId, wwPassport.River, wwPassport.Title)
			}
			switch len(wwpts) {
			case 1:
				wwPassport.WwId = wwpts[0].Id
				err = this.WwPassportDao.Upsert(wwPassport)
				if err != nil {
					log.Fatal("Can not upsert ww passport", wwPassport, err)
				}
				this.findAndStoreImages(wwPassport, catalogConnector)
			case 0:
				log.Warn("No ww point found for %s - %s", wwPassport.River, wwPassport.Title)
				continue
			default:
				log.Warn("Can not explicitly detect ww point for ww passport", wwPassport.Source, wwPassport.RemoteId, wwPassport.River, wwPassport.River)
				continue
			}
		case 0:
			log.Warn("No river found", wwPassport.River)
			continue
		default:
			log.Warn("Can not explicitly detect river for ww passport", wwPassport.Source, wwPassport.RemoteId, wwPassport.River)
			continue
		}
	}
}

func (this App) findAndStoreImages(wwPassport dao.WWPassport, catalogConnector common.CatalogConnector) {
	log.Infof("Find images for ww passport %s-%s", wwPassport.Source, wwPassport.RemoteId)
	imgs, err := catalogConnector.GetImages(wwPassport.RemoteId)
	if err != nil {
		this.Fatalf(err, "Can not load images for ww passport %s-%s %d", wwPassport.Source, wwPassport.RemoteId, wwPassport.WwId)
	}
	_, err = this.ImgDao.Upsert(imgs...)
	if err != nil {
		this.Fatalf(err, "Can not upsert images for report %d", wwPassport.WwId)
	}
}

func (this App) DoSyncReports() {
	lastId, err := this.VoyageReportDao.GetLastId(huskytm.SOURCE)
	if err != nil {
		this.Fatalf(err, "Can not connect get last report id")
	}
	lastReportIdStr := lastId.(time.Time).Format(huskytm.TIME_FORMAT)
	log.Infof("Get and store reports since %s", lastReportIdStr)

	reportProvider, err := huskytm.GetReportProvider(this.Configuration.Login, this.Configuration.Password)
	if err != nil {
		this.Fatalf(err, "Can not connect to source")
	}
	defer reportProvider.Close()

	reports, next, err := reportProvider.ReportsSince(lastReportIdStr)
	if err != nil {
		log.Fatal(err, "Can not get posts")
	}
	if len(reports) == 0 {
		next = lastReportIdStr
	}

	reports, err = this.VoyageReportDao.UpsertVoyageReports(reports...)
	if err != nil {
		this.Fatalf(err, "Can not store reports: %v", reports)
	}

	log.Infof("%d reports are successfully stored. Next id is %s\n", len(reports), next)

	log.Info("Try to connect reports with known rivers")

	for i, report := range reports {
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
		reports[i].Rivers = rivers
	}

	for _, report := range reports {
		this.findMatchAndStoreImages(report, reportProvider)
	}
}

func (this App) findMatchAndStoreImages(report dao.VoyageReport, reportProvider common.ReportProvider) {
	log.Infof("Find images for report %d: %s %s", report.Id, report.RemoteId, report.Title)
	imgs, err := reportProvider.Images(report.RemoteId)
	if err != nil {
		this.Fatalf(err, "Can not load images for report %d", report.Id)
	}
	fmt.Printf("%d images found\n", len(imgs))

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

	log.Infof("Store %d images for report %d", len(matchedImgs), report.Id)
	_, err = this.ImgDao.Upsert(matchedImgs...)
	if err != nil {
		this.Fatalf(err, "Can not upsert images for report %d", report.Id, )
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
				this.Fatalf(err, "Can not list white water spots for river %d", river.Id)
			}
			for _, wwpt := range wwpts {
				for _, label := range img.LabelsForSearch {
					if strings.Contains(forCompare(label), forCompare(wwpt.Title)) {
						fmt.Println("Found: ", label)
						rec, found := candidates[img.RemoteId]
						if !found {
							rec = ImgWwPoints{
								Img:img,
								Wwpts:[]dao.WhiteWaterPointWithRiverTitle{},
							}
						}
						rec.Wwpts = append(rec.Wwpts, wwpt)
						candidates[img.RemoteId] = rec
					}
				}
			}
		}
	}
	return candidates
}

func forCompare(s string) string {
	return strings.Replace(strings.ToLower(s), "ё", "e", -1)
}

func (this App) Fatalf(err error, pattern string, args ...interface{}) {
	this.Report(err)
	log.Fatalf(pattern + ": " + err.Error(), args)
}

func (this App) Report(err error) {
	templateData, err := emailTemplateBytes()
	if err != nil {
		log.Fatal("Can not load email template:\t", err)
	}

	tmpl, err := template.New("report-email").Parse(string(templateData))
	if err != nil {
		log.Fatal("Can not compile email template:\t", err)
	}

	err = mail.SendMail(this.Notifications.EmailSender, this.Notifications.EmailRecipients, this.Notifications.ImportExportEmailSubject, func(w io.Writer) error {
		return tmpl.Execute(w, *this.stat)
	})
}