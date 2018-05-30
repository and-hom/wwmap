package common

import (
	"github.com/and-hom/wwmap/lib/dao"
	"io"
)

type ReportProvider interface {
	io.Closer
	ReportsSince(key string) ([]dao.VoyageReport, string, error);
	Images(reportId string) ([]dao.Img, error);
}

type CatalogConnector interface {
	io.Closer
	PassportIdsSince(key string) []string
	GetPassport(key string) dao.WhiteWaterPoint
	GetImages(key string) []dao.Img

	Exists(key string) bool
	Create(passport dao.WWPassport, imgs []dao.Img)
}
