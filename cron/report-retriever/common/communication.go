package common

import (
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/sirupsen/logrus"
	"io"
	"time"
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

func (this WithReportProvider)  SourceId() string {
	provider, err := this()
	if err != nil {
		logrus.Errorf("Can not create connector: %v", err)
		return ""
	}
	return provider.SourceId()
}
