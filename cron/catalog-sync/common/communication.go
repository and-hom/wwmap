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
		if provider != nil {
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
	Category string
}

type CountryLink struct {
	LinkOnPage
	Code string
}

type VoyageReportLink struct {
	LinkOnPage
	SourceLogo string
}

type SpotPageDto struct {
	Id          int

	Spot            dao.WhiteWaterPointFull
	River           dao.River
	Region          dao.Region
	Country         dao.Country

	MainImg         dao.Img
	Imgs            []dao.Img

	RootPageLink    string
	CountryPageLink string
	RegionPageLink  string
	RiverPageLink   string
}
type RiverPageDto struct {
	Id          int

	River           dao.River
	Region          dao.Region
	Country         dao.Country

	Links           []SpotLink
	MainImg         dao.Img
	Reports         []VoyageReportLink

	RootPageLink    string
	CountryPageLink string
	RegionPageLink  string
}

type RegionPageDto struct {
	Id          int

	Region          dao.Region
	Country         dao.Country

	Links           []LinkOnPage

	RootPageLink    string
	CountryPageLink string
}

type CountryPageDto struct {
	Id       int

	Country      dao.Country

	RegionLinks  []LinkOnPage
	RiverLinks   []LinkOnPage

	RootPageLink string
}

type RootPageDto struct {
	Id       int
	Links []CountryLink
}

type CatalogConnector interface {
	io.Closer
	PassportEntriesSince(key string) ([]dao.WWPassport, error)
	GetImages(key string) ([]dao.Img, error)

	CreateEmptyPageIfNotExistsAndReturnId(parent int, pageId int, title string) (int, string, bool, error)
	WriteSpotPage(SpotPageDto) error
	WriteRiverPage(RiverPageDto) error
	WriteRegionPage(RegionPageDto) error
	WriteCountryPage(CountryPageDto) error
	WriteRootPage(RootPageDto) error
}
