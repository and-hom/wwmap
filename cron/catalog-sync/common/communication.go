package common

import (
	"github.com/and-hom/wwmap/lib/dao"
	"io"
	"github.com/and-hom/wwmap/lib/model"
)

type ReportProvider interface {
	io.Closer
	ReportsSince(key string) ([]model.VoyageReport, string, error);
	Images(reportId int) ([]model.Img, error);
}

type CatalogConnector interface {
	io.Closer
	PassportIdsSince(key string) []string
	GetPassport(key string) dao.WhiteWaterPoint
	GetImages(key string) []model.Img

	Exists(key string) bool
	Create(passport model.WWPassport, imgs []model.Img)
}
