package common

import (
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/sirupsen/logrus"
	"io"
)

type WithCatalogConnector struct {
	F      func() (CatalogConnector, error)
	cached CatalogConnector
}

func (this WithCatalogConnector) getConnector() (CatalogConnector, error) {
	var connector CatalogConnector
	var err error

	if this.cached == nil {
		connector, err = this.F()
	} else {
		connector = this.cached
	}
	if err != nil {
		if connector != nil {
			return nil, fmt.Errorf("Can not connect to source %s: %v", connector.SourceId(), err)
		} else {
			return nil, fmt.Errorf("Can not connect to source unknown (nil provider): %v", err)
		}
	}
	return connector, err
}
func (this WithCatalogConnector) Do(payload func(CatalogConnector) error) error {
	connector, err := this.getConnector()
	if err != nil {
		return err
	}
	return payload(connector)
}

func (this WithCatalogConnector) SourceId() string {
	connector, err := this.getConnector()
	if err != nil {
		logrus.Errorf("Can not create connector: %v", err)
		return ""
	}
	return connector.SourceId()
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
	Id              int

	Spot            dao.WhiteWaterPointFull
	River           dao.River
	Region          dao.Region
	Country         dao.Country

	MainImage       dao.Img
	Imgs            []dao.Img
	Videos          []dao.Img

	RootPageLink    string
	CountryPageLink string
	RegionPageLink  string
	RiverPageLink   string
}
type RiverPageDto struct {
	Id              int

	River           dao.River
	Region          dao.Region
	Country         dao.Country

	Links           []SpotLink
	MainImage       dao.Img
	Reports         []VoyageReportLink

	RootPageLink    string
	CountryPageLink string
	RegionPageLink  string
}

type RegionPageDto struct {
	Id              int

	Region          dao.Region
	Country         dao.Country

	Links           []LinkOnPage

	RootPageLink    string
	CountryPageLink string
}

type CountryPageDto struct {
	Id           int

	Country      dao.Country

	RegionLinks  []LinkOnPage
	RiverLinks   []LinkOnPage

	RootPageLink string
}

type RootPageDto struct {
	Id    int
	Links []CountryLink
}

type CatalogConnector interface {
	io.Closer
	SourceId() string
	FailOnFirstError() bool

	CreateEmptyPageIfNotExistsAndReturnId(id int64, parent int, pageId int, title string) (int, string, bool, error)
	WriteSpotPage(SpotPageDto) error
	WriteRiverPage(RiverPageDto) error
	WriteRegionPage(RegionPageDto) error
	WriteCountryPage(CountryPageDto) error
	WriteRootPage(RootPageDto) error
}
