package common

import (
	"github.com/and-hom/wwmap/lib/dao"
	"io"
	"time"
	"fmt"
)

type ReportProvider interface {
	io.Closer
	SourceId() string
	ReportsSince(t time.Time) ([]dao.VoyageReport, time.Time, error);
	Images(reportId string) ([]dao.Img, error);
}

type WithReportProvider func() (ReportProvider, error)

func (this WithReportProvider) Do(payload func(ReportProvider) error) error {
	provider, err := this()
	if err != nil {
		if provider!=nil {
			return fmt.Errorf("Can not connect to source %s: %v", provider.SourceId(), err)
		} else {
			return fmt.Errorf("Can not connect to source unknown (nil provider): %v", err)
		}
	}
	defer provider.Close()

	return payload(provider)
}

type LinkOnPage struct {
	Url   string
	Title string
}

type SpotLink struct {
	LinkOnPage
	Category    string
}

type CountryLink struct {
	LinkOnPage
	Code string
}

type VoyageReportLink struct {
	LinkOnPage
	SourceLogo string
}

type CatalogConnector interface {
	io.Closer
	PassportEntriesSince(key string) ([]dao.WWPassport, error)
	GetImages(key string) ([]dao.Img, error)

	CreateEmptyPageIfNotExistsAndReturnId(parent int, pageId int, title string) (int, string, bool, error)
	WriteSpotPage(pageId int, spot dao.WhiteWaterPointFull, river dao.River, region dao.Region, country dao.Country, mainImg dao.Img, imgs []dao.Img, rootPageLink, countryPageLink, regionPageLink, riverPageLink string) error
	WriteRiverPage(pageId int, river dao.River, region dao.Region, country dao.Country, links []SpotLink, rootPageLink, countryPageLink, regionPageLink string, mainImg dao.Img, reports []VoyageReportLink) error
	WriteRegionPage(pageId int, region dao.Region, country dao.Country, links []LinkOnPage, rootPageLink, countryPageLink string) error
	WriteCountryPage(pageId int, country dao.Country, regionLinks, riverLinks []LinkOnPage, rootPageLink string) error
	WriteRootPage(pageId int, countryLinks []CountryLink) error
}
