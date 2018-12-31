package huskytm

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	wp "github.com/and-hom/go-wordpress"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm/templates"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"net/http"
	"time"
)

func GetCatalogConnector(login, password string, minDeltaBetweenRequests time.Duration) (common.CatalogConnector, error) {
	client := wp.NewClient(&wp.Options{
		BaseAPIURL: API_BASE, // example: `http://192.168.99.100:32777/wp-json/wp/v2`
		Username:   login,
		Password:   password,
		Timeout: 10 * time.Second,
	})
	u, r, b, e := client.Users().Me(emptyMap())
	if e != nil {
		return nil, fmt.Errorf("Connection failed: %s", e.Error())
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Connection failed. Code: %d Body: %s", r.StatusCode, string(b))
	}
	t, err := common.LoadTemplates(templates.MustAsset)
	if err!=nil {
		return nil, err
	}
	return &HuskytmCatalogConnector{
		client:client,
		me:u.ID,
		pageIdsCache:make(map[string]int),
		templates:t,
		rateLimit:util.NewRateLimit(minDeltaBetweenRequests),
	}, nil
}

type HuskytmCatalogConnector struct {
	client       *wp.Client
	me           int
	pageIdsCache map[string]int
	templates    common.Templates
	rateLimit    util.RateLimit
}

func (this *HuskytmCatalogConnector) SourceId() string {
	return SOURCE
}

func (this *HuskytmCatalogConnector) Close() error {
	return nil
}

func (this *HuskytmCatalogConnector) CreateEmptyPageIfNotExistsAndReturnId(id int64, parent int, pageId int, title string) (int, string, bool, error) {
	log.Infof("Check page for id=%d page_id=%d", id, pageId)
	if pageId <= 0 {
		log.Infof("Really create new page for %s", title)
		createdPageId, link, err := this.createPage(parent, title)
		created := err == nil
		return createdPageId, link, created, err
	} else {
		log.Infof("Page exists: %d %s - do check", pageId, title)
	}
	this.rateLimit.WaitIfNecessary()
	p, r, _, err := this.client.Pages().Get(pageId, emptyMap())
	if r.StatusCode == http.StatusNotFound || p.Status == "trash" {
		log.Warnf("Existing page %d is not sutable: %d %s", pageId, r.StatusCode, p.Status)
		createdPageId, link, err := this.createPage(parent, title)
		log.Info("Created page id=%d for entity id=%d", createdPageId, id)
		created := err == nil
		return createdPageId, link, created, err
	} else if err != nil {
		return 0, "", false, err
	}
	return p.ID, p.Link, false, nil
}

func (this *HuskytmCatalogConnector) createPage(parent int, title string) (int, string, error) {
	this.rateLimit.WaitIfNecessary()
	p, r, b, err := this.client.Pages().Create(&wp.Page{
		Title:        wp.Title{Raw:title},
		Author:        this.me,
		Parent:        parent,
		Status:        "publish",
	})
	if err != nil {
		log.Errorf("Connection failed. Code: %d Body: %s", r.StatusCode, string(b))
		return 0, "", err
	}
	return p.ID, p.Link, nil
}

func (this *HuskytmCatalogConnector) WriteSpotPage(page common.SpotPageDto) error {
	return this.writePage(page.Id, this.templates.WriteSpot, page.Spot.Title, page)
}
func (this *HuskytmCatalogConnector) WriteRiverPage(page common.RiverPageDto) error {
	return this.writePage(page.Id, this.templates.WriteRiver, page.River.Title, page)
}
func (this *HuskytmCatalogConnector) WriteRegionPage(page common.RegionPageDto) error {
	return this.writePage(page.Id, this.templates.WriteRegion, page.Region.Title, page)
}
func (this *HuskytmCatalogConnector) WriteCountryPage(page common.CountryPageDto) error {
	return this.writePage(page.Id, this.templates.WriteCountry, page.Country.Title, page)
}

func (this *HuskytmCatalogConnector) WriteRootPage(page common.RootPageDto) error {
	return this.writePage(page.Id, this.templates.WriteRoot, "Каталог водных препятствий", page)
}

func (this *HuskytmCatalogConnector) writePage(pageId int, tmpl func(data interface{}) (string, error), title string, data interface{}) error {
	log.Infof("Write page %d for %s", pageId, title)

	body, err := tmpl(data)
	if err != nil {
		log.Errorf("Can not process template", err)
		return err
	}

	page := wp.Page{
		Title:  wp.Title{Raw:title},
		Author:        this.me,
		Content:wp.Content{Raw:body},
	}

	this.rateLimit.WaitIfNecessary()
	_, r, b, err := this.client.Pages().Update(pageId, &page)
	if err != nil {
		bodyStr := ""
		if b != nil {
			bodyStr = string(b)
		}
		statusCode := 0
		if r!=nil {
			statusCode = r.StatusCode
		}

		log.Errorf("Connection failed. Code: %d Body: %s", statusCode, bodyStr)
		return err
	}

	return nil
}

func (this *HuskytmCatalogConnector) PassportEntriesSince(key string) ([]dao.WWPassport, error) {
	return []dao.WWPassport{}, nil
}
func (this *HuskytmCatalogConnector) GetPassport(key string) (dao.WhiteWaterPoint, error) {
	return dao.WhiteWaterPoint{}, nil
}
func (this *HuskytmCatalogConnector) GetImages(key string) ([]dao.Img, error) {
	return []dao.Img{}, nil
}

type PageNotFoundError struct {
	msg string
}

func (this PageNotFoundError) Error() string {
	return this.msg
}

