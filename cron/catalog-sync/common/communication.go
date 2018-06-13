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
	PassportEntriesSince(key string) ([]dao.WWPassport, error)
	GetImages(key string) ([]dao.Img, error)

	Exists(key []string) (bool, error)
	CreatePage(title string, parent int) (int, error)
	GetId(title string, parent int) (int, error)
	Create(passport dao.WhiteWaterPoint, parent int, imgs []dao.Img) error
}
