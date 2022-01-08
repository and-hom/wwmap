package pdf

//go:generate go-bindata -pkg $GOPACKAGE -o templates.go -prefix templates/  ./templates

import (
	"context"
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/and-hom/wwmap/cron/catalog-export/common"
	"github.com/and-hom/wwmap/lib/blob"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

const SOURCE = "pdf"
const MAX_PDF_GENERATION_DURATION = 20 * time.Second

func GetCatalogConnector(pdfStorage, htmlStorage blob.BlobStorage, pageLinkTemplate string) (common.CatalogConnector, error) {
	t, err := common.LoadTemplates(MustAsset)
	if err != nil {
		return nil, err
	}

	return &PdfCatalogConnector{
		templates:        t,
		pdfStorage:       pdfStorage,
		htmlStorage:      htmlStorage,
		spotBuf:          []common.SpotPageDto{},
		pageLinkTemplate: pageLinkTemplate,
	}, nil
}

type PdfCatalogConnector struct {
	pdfStorage       blob.BlobStorage
	htmlStorage      blob.BlobStorage
	templates        common.Templates
	spotBuf          []common.SpotPageDto
	pageLinkTemplate string
}

func (this *PdfCatalogConnector) SourceId() string {
	return SOURCE
}

func (this *PdfCatalogConnector) FailOnFirstError() bool {
	return false
}

func (this *PdfCatalogConnector) Close() error {
	return nil
}

func (this *PdfCatalogConnector) CreateEmptyPageIfNotExistsAndReturnId(id int64, parent int, pageId int, title string) (int, string, bool, error) {
	return int(id), fmt.Sprintf(this.pageLinkTemplate, id), true, nil
}

func (this *PdfCatalogConnector) WriteSpotPage(page common.SpotPageDto) error {
	this.spotBuf = append(this.spotBuf, page)
	return nil
}
func (this *PdfCatalogConnector) WriteRiverPage(page common.RiverPageDto) error {
	b, err := this.templates.WriteRiver(RiverPageDto{RiverPageDto: page, Spots: this.spotBuf})
	if err != nil {
		log.Error("Can not process template", err)
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
	log.Infof("Write html page %d for %s", pageId, title)
	htmlFileKey := fmt.Sprintf("%d.htm", pageId)
	if err := this.htmlStorage.Remove(htmlFileKey); err!=nil && !os.IsNotExist(err) {
		log.Error("Can't write html page: ", err)
		return err
	}
	if err := this.htmlStorage.Store(htmlFileKey, strings.NewReader(body)); err!=nil {
		log.Error("Can't write html page: ", err)
		return err
	}

	log.Infof("Write pdf page %d for %s", pageId, title)
	pdfFileKey := fmt.Sprintf("%d.pdf", pageId)
	if err := this.htmlStorage.Remove(pdfFileKey); err!=nil && !os.IsNotExist(err) {
		log.Error("Can't write pdf page: ", err)
		return err
	}

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
	pdfGenerator.Title.Set(title)

	ctx, cancelFunc := context.WithTimeout(context.TODO(), MAX_PDF_GENERATION_DURATION)
	defer cancelFunc()

	cacheDir := os.TempDir() + "/wwmap-wkhtml-cache"
	os.MkdirAll(cacheDir, os.ModePerm)

	pr := wkhtmltopdf.NewPageReader(strings.NewReader(body))
	pr.NoStopSlowScripts.Set(true)
	pr.WindowStatus.Set("LOAD_FINISHED")
	pr.CacheDir.Set(cacheDir)
	pr.LoadErrorHandling.Set("ignore")
	pr.LoadMediaErrorHandling.Set("ignore")
	pdfGenerator.AddPage(pr)

	err = pdfGenerator.CreateContext(ctx)
	if err != nil {
		log.Errorf("Can not render pdf - remove if exists: %v", err)
		err2 := this.pdfStorage.Remove(pdfFileKey)
		if err2!=nil {
			log.Warnf("Can not remove: %v", err2)
		}
		return nil
	}

	err = this.pdfStorage.Store(pdfFileKey, pdfGenerator.Buffer())
	if err != nil {
		log.Errorf("Can not write file: %v", err)
		return err
	}

	log.Infof("Page %d for %s written successfully!", pageId, title)
	return nil
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
