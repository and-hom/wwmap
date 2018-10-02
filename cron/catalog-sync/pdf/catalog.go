package pdf

import (
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"github.com/and-hom/wwmap/cron/catalog-sync/pdf/templates"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/blob"
	"strings"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

const SOURCE = "pdf"

func GetCatalogConnector(storage blob.BlobStorage) (common.CatalogConnector, error) {
	t, err := common.LoadTemplates(templates.MustAsset)
	if err != nil {
		return nil, err
	}

	return &PdfCatalogConnector{
		templates:t,
		storage:storage,
		spotBuf:[]common.SpotPageDto{},
	}, nil
}

type PdfCatalogConnector struct {
	storage   blob.BlobStorage
	templates common.Templates
	spotBuf   []common.SpotPageDto
}

func (this *PdfCatalogConnector) SourceId() string {
	return SOURCE
}

func (this *PdfCatalogConnector) Close() error {
	return nil
}

func (this *PdfCatalogConnector) CreateEmptyPageIfNotExistsAndReturnId(id int64, parent int, pageId int, title string) (int, string, bool, error) {
	return int(id), "pdf", true, nil
}

func (this *PdfCatalogConnector) WriteSpotPage(page common.SpotPageDto) error {
	this.spotBuf = append(this.spotBuf, page)
	return nil
}
func (this *PdfCatalogConnector) WriteRiverPage(page common.RiverPageDto) error {
	b, err := this.templates.WriteRiver(RiverPageDto{RiverPageDto: page, Spots:this.spotBuf, })
	if err != nil {
		log.Errorf("Can not process template", err)
		return err
	}
	err = this.writePage(page.Id, b, page.River.Title)
	this.spotBuf = []common.SpotPageDto{}
	return err
}
func (this *PdfCatalogConnector) WriteRegionPage(page common.RegionPageDto) error {
	return nil
}
func (this *PdfCatalogConnector) WriteCountryPage(page common.CountryPageDto) error {
	return nil
}

func (this *PdfCatalogConnector) WriteRootPage(page common.RootPageDto) error {
	return nil
}

func (this *PdfCatalogConnector) writePage(pageId int, body string, title string) error {
	log.Infof("Write page %d for %s", pageId, title)
	pdfGenerator, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		log.Error("Can not create PDF generator")
		return err
	}
	pdfGenerator.Dpi.Set(300)
	pdfGenerator.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfGenerator.PageSize.Set("A4")
	pdfGenerator.MarginTop.Set(15)
	pdfGenerator.MarginBottom.Set(20)
	pdfGenerator.MarginLeft.Set(10)
	pdfGenerator.MarginRight.Set(10)

	pr := wkhtmltopdf.NewPageReader(strings.NewReader(body))
	pdfGenerator.AddPage(pr)

	err = pdfGenerator.Create()
	if err != nil {
		log.Errorf("Can not render pdf", err)
		return err
	}

	err = this.storage.Store(fmt.Sprintf("%d.pdf", pageId), pdfGenerator.Buffer())
	if err != nil {
		log.Errorf("Can not write file", err)
		return err
	}

	return nil
}

func (this *PdfCatalogConnector) PassportEntriesSince(key string) ([]dao.WWPassport, error) {
	return []dao.WWPassport{}, nil
}
func (this *PdfCatalogConnector) GetPassport(key string) (dao.WhiteWaterPoint, error) {
	return dao.WhiteWaterPoint{}, nil
}
func (this *PdfCatalogConnector) GetImages(key string) ([]dao.Img, error) {
	return []dao.Img{}, nil
}

type PageNotFoundError struct {
	msg string
}

func (this PageNotFoundError) Error() string {
	return this.msg
}

type RiverPageDto struct {
	common.RiverPageDto
	Spots []common.SpotPageDto
}

