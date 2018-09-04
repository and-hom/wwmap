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
		return fmt.Errorf("Can not connect to source %s: %s", provider.SourceId(), err.Error())
	}
	defer provider.Close()

	return payload(provider)
}

type CatalogConnector interface {
	io.Closer
	PassportEntriesSince(key string) ([]dao.WWPassport, error)
	GetImages(key string) ([]dao.Img, error)

	Exists(key []string) (bool, error)
	CreatePage(title string, parent int) (int, error)
	GetId(title string, parent int) (int, error)
	Create(passport dao.WhiteWaterPointFull, parent int, mainImage dao.Img, imgs []dao.Img) error
}
