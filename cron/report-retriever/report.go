package main

//go:generate go-bindata -pkg $GOPACKAGE -o bindata.go ./email-template

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/report-retriever/common"
	"github.com/and-hom/wwmap/cron/vodinfo-eye/graduation"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"html/template"
	"net/http"
	"strings"
	"time"
)

const COMMON_TIME_FORMAT string = "2006-01-02T15:04:05"

func (this *App) DoSyncReports() {
	for _, rpf := range this.reportProviders {
		err := rpf.Do(func(rp common.ReportProvider) error {
			if this.sourceOnly == "" || this.sourceOnly == rp.SourceId() {
				return this.doSyncReports(&rp)
			} else {
				return nil
			}
		})
		if err != nil {
			log.Errorf("Can not access to source: %v", err)
		}
	}
}

func (this *App) doSyncReports(reportProvider *common.ReportProvider) error {
	source := (*reportProvider).SourceId()
	lastId, err := this.VoyageReportDao.GetLastId(source)
	if err != nil {
		log.Error("Can not connect get last report id")
		return err
	}
	log.Infof("Get and store reports from %s since %s", source, lastId.(time.Time).Format(COMMON_TIME_FORMAT))

	reports, next, err := (*reportProvider).ReportsSince(lastId.(time.Time))
	if err != nil {
		log.Error("Can not get posts: ", err)
		return err
	}
	if len(reports) == 0 {
		next = lastId.(time.Time)
	}

	reports, err = this.VoyageReportDao.UpsertVoyageReports(reports...)
	if err != nil {
		log.Errorf("Can not store reports from %s: %v", source, reports)
		return err
	}

	log.Infof("%d reports from %s are successfully stored. Next id is %s\n", len(reports), source, next)

	reportsToRivers := make(map[int64][]dao.RiverTitle)
	for i := 0; i < len(reports); i++ {
		reportsToRivers[reports[i].Id] = make([]dao.RiverTitle, 0)
	}

	log.Info("Try to connect reports with known rivers")
	err = this.associateReportsWithRivers(source, &reportsToRivers)
	if err != nil {
		log.Error("Can not associate rivers with reports: ", err)
		return err
	}

	for _, report := range reports {
		rivers, found := reportsToRivers[report.Id]
		if !found {
			rivers = []dao.RiverTitle{}
		}
		err := this.findMatchAndStoreImages(report, rivers, reportProvider)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *App) associateReportsWithRivers(source string, resultHandlerMap *map[int64][]dao.RiverTitle) error {
	return this.VoyageReportDao.ForEach(source, func(report *dao.VoyageReport) error {
		return this.associateReportWithRiver(report, resultHandlerMap)
	})
}

func (this *App) associateReportWithRiver(report *dao.VoyageReport, resultHandlerMap *map[int64][]dao.RiverTitle) error {
	log.Debugf("Tags are: %v", report.Tags)
	rivers, err := this.RiverDao.FindTitles(report.Tags)
	if err != nil {
		log.Error("Can not find rivers by tags")
		return err
	}
	for _, river := range rivers {
		err := this.VoyageReportDao.AssociateWithRiver(report.Id, river.Id)
		if err != nil {
			return err
		}
		riversForReport, found := (*resultHandlerMap)[report.Id]
		if found {
			(*resultHandlerMap)[report.Id] = append(riversForReport, river)
		}
	}
	return nil
}

func (this *App) findMatchAndStoreImages(report dao.VoyageReport, rivers []dao.RiverTitle, reportProvider *common.ReportProvider) error {
	log.Infof("Find images for report %d: %s %s", report.Id, report.RemoteId, report.Title)
	imgs, err := (*reportProvider).Images(report.RemoteId)
	if err != nil {
		log.Errorf("Can not load images for report %d", report.Id)
		return err
	}
	log.Infof("%d images found for %s %d", len(imgs), report.Source, report.Id)
	log.Infof("Bind images to ww spots for report %d", report.Id)
	matchedImgs := []dao.Img{}
	candidates, err := this.matchImgsToWhiteWaterPoints(report, imgs, rivers)
	if err != nil {
		return err
	}
	log.Infof("%d images matched for %s %d", len(candidates), report.Source, report.Id)

	for _, imgToWwpts := range candidates {
		if len(imgToWwpts.Wwpts) == 1 {
			imgToWwpts.Img.WwId = imgToWwpts.Wwpts[0].Id
			matchedImgs = append(matchedImgs, imgToWwpts.Img)
		} else if len(imgToWwpts.Wwpts) > 1 {
			log.Warn("More then one white water point for img ", imgToWwpts.Img.RemoteId)
		}
	}

	client := http.Client{Timeout: 30 * time.Second}

	for _, img := range matchedImgs {
		this.rateLimit.WaitIfNecessary()
		log.Infof("Fetch image %s to get date from exif", img.RawUrl)
		resp, err := client.Get(img.RawUrl)
		if err != nil {
			log.Error("Can't download image for exif parsing: ", err)
			continue
		}
		defer resp.Body.Close()

		imgDate := util.GetImageRealDate(resp.Body)

		if imgDate != nil {
			img.Date = imgDate
			img.DateLevelUpdated = *imgDate
			img.Level = this.getLevelsForDate(img.WwId, imgDate)
		}
	}

	log.Infof("Store %d images for report %d", len(matchedImgs), report.Id)
	_, err = this.ImgDao.Upsert(matchedImgs...)
	if err != nil {
		log.Errorf("Can not upsert images for report %d", report.Id)
	}
	return err
}

func (this *App) getLevelsForDate(spotId int64, date *time.Time) map[string]int8 {
	if date == nil {
		return make(map[string]int8)
	}
	river, err := this.RiverDao.FindForSpot(spotId)
	if err != nil {
		log.Errorf("Can not select river for spot: id=%d", spotId)
		return make(map[string]int8)
	} else {
		sensorIds := river.GetSensorIds()
		return graduation.GetLevelBySensors(this.LevelSensorDao, this.LevelDao, sensorIds, *date, 1, -1)
	}
}

type ImgWwPoints struct {
	Img   dao.Img
	Wwpts []dao.WhiteWaterPointWithRiverTitle
}

func (this *App) matchImgsToWhiteWaterPoints(report dao.VoyageReport, imgs []dao.Img, rivers []dao.RiverTitle) (map[string]ImgWwPoints, error) {
	candidates := make(map[string]ImgWwPoints)
	for _, img := range imgs {
		for _, river := range rivers {
			wwpts, err := this.WhiteWaterDao.ListByRiver(river.Id)
			if err != nil {
				log.Errorf("Can not list white water spots for river %d", river.Id)
				return candidates, err
			}
			for _, wwpt := range wwpts {
				for _, label := range img.LabelsForSearch {
					if strings.Contains(forCompare(label), forCompare(wwpt.Title)) {
						log.Info("Found: ", label)
						rec, found := candidates[img.RemoteId]
						img.ReportId = report.Id
						if !found {
							rec = ImgWwPoints{
								Img:   img,
								Wwpts: []dao.WhiteWaterPointWithRiverTitle{},
							}
						}
						rec.Wwpts = append(rec.Wwpts, wwpt)
						candidates[img.RemoteId] = rec
					}
				}
			}
		}
	}
	return candidates, nil
}

func forCompare(s string) string {
	return strings.Replace(strings.Replace(strings.ToLower(s), "ё", "e", -1), "-", " ", -1)
}

func (this *App) Fatalf(err error, pattern string, args ...interface{}) {
	this.Report(err)
	log.Fatalf(pattern+": "+err.Error(), args)
}

func (this *App) Report(err error) {
	tmpl, err := template.New("report-email").Parse(string(MustAsset("email-template")))
	if err != nil {
		log.Fatal("Can not compile email template:\t", err)
	}
	buf := bytes.Buffer{}
	if err = tmpl.Execute(&buf, *this.stat); err != nil {
		log.Fatal("Can not process email template:\t", err)
	}
	this.NotificationHelper.SendToRole(dao.Notification{
		Comment:    buf.String(),
		Recipient:  dao.NotificationRecipient{Provider: dao.NOTIFICATION_PROVIDER_EMAIL, Recipient: "info@wwmap.ru"},
		Classifier: "repor-import",
		SendBefore: time.Now().Add(2 * time.Hour),
	}, dao.ADMIN)
}
