package main

import (
	"github.com/and-hom/wwmap/lib/dao"
	"strings"
	log "github.com/Sirupsen/logrus"
	"io"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm"
	"time"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"fmt"
	"html/template"
	"github.com/and-hom/wwmap/lib/mail"
)

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
			wwpts, err := this.WhiteWaterDao.ListByRiver(river.Id)
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
	return strings.Replace(strings.Replace(strings.ToLower(s), "ё", "e", -1), "-", " ", -1)
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
