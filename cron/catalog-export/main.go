package main

import (
	"flag"
	"fmt"
	"github.com/and-hom/wwmap/cron/catalog-export/common"
	"github.com/and-hom/wwmap/cron/catalog-export/huskytm"
	"github.com/and-hom/wwmap/cron/catalog-export/pdf"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	log "github.com/sirupsen/logrus"
	"strings"
)

type App struct {
	CountryDao      dao.CountryDao
	RegionDao       dao.RegionDao
	RiverDao        dao.RiverDao
	WhiteWaterDao   dao.WhiteWaterDao
	ImgDao          dao.ImgDao
	VoyageReportDao dao.VoyageReportDao
	NotificationDao dao.NotificationDao

	Configuration config.WordpressSync

	ImgUrlBase        string
	ImgUrlPreviewBase string
	ResourceBase      string

	catalogConnectors []common.WithCatalogConnector

	sourceOnly string
}

func CreateApp() App {
	configuration := config.Load("")
	configuration.ConfigureLogger()

	pgStorage := dao.NewPostgresStorage(configuration.Db)
	riverPassportPdfStorage := blob.BasicFsStorage{
		BaseDir: configuration.RiverPassportPdfStorage.Dir,
	}
	riverPassportHtmlStorage := blob.BasicFsStorage{
		BaseDir: configuration.RiverPassportHtmlStorage.Dir,
	}
	notificationDao := dao.NewNotificationPostgresDao(pgStorage)
	return App{
		VoyageReportDao: dao.NewVoyageReportPostgresDao(pgStorage),
		CountryDao:      dao.NewCountryPostgresDao(pgStorage),
		RegionDao:       dao.NewRegionPostgresDao(pgStorage),
		RiverDao:        dao.NewRiverPostgresDao(pgStorage),
		WhiteWaterDao:   dao.NewWhiteWaterPostgresDao(pgStorage),
		ImgDao:          dao.NewImgPostgresDao(pgStorage),
		NotificationDao: notificationDao,
		Configuration:   configuration.Sync,
		catalogConnectors: []common.WithCatalogConnector{
			{F: func() (common.CatalogConnector, error) {
				return huskytm.GetCatalogConnector(configuration.Sync.Login, configuration.Sync.Password, configuration.Sync.MinDeltaBetweenRequests)
			}},
			{F: func() (common.CatalogConnector, error) {
				return pdf.GetCatalogConnector(riverPassportPdfStorage, riverPassportHtmlStorage, configuration.RiverPassportPdfStorage.UrlBase)
			}},
		},
		ImgUrlBase:        configuration.ImgStorage.Full.UrlBase,
		ImgUrlPreviewBase: configuration.ImgStorage.Preview.UrlBase,
		ResourceBase:      configuration.Content.ResourceBase,
	}
}

func parseFlags(app App) (string, error) {
	sourceIdsMap := make(map[string]bool)
	sourceIds := []string{}
	for _, c := range app.catalogConnectors {
		sourceIdsMap[c.SourceId()] = true
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
	log.Infof("Starting catalog export")
	app := CreateApp()

	source, err := parseFlags(app)
	if err != nil {
		log.Fatal(err)
	}

	app.sourceOnly = source
	log.Infof("Write catalog to %s", source)
	app.DoWriteCatalog()
}
