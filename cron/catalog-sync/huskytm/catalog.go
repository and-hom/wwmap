package huskytm

import (
	"github.com/and-hom/wwmap/lib/dao"
	wp "github.com/and-hom/go-wordpress"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm/templates"
	"fmt"
	"net/http"
	"html/template"
	log "github.com/Sirupsen/logrus"
	"bytes"
	"time"
	"github.com/and-hom/wwmap/lib/util"
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
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
	}

	spotPageTemplate, e := template.New("spot").Funcs(funcMap).Parse(string(templates.MustAsset("spot-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile template: %s", e.Error())
	}
	riverPageTemplate, e := template.New("river").Funcs(funcMap).Parse(string(templates.MustAsset("river-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile template: %s", e.Error())
	}
	regionPageTemplate, e := template.New("region").Funcs(funcMap).Parse(string(templates.MustAsset("region-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile template: %s", e.Error())
	}
	countryPageTemplate, e := template.New("country").Funcs(funcMap).Parse(string(templates.MustAsset("country-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile template: %s", e.Error())
	}
	rootPageTemplate, e := template.New("root").Funcs(funcMap).Parse(string(templates.MustAsset("root-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile template: %s", e.Error())
	}
	return &HuskytmCatalogConnector{
		client:client,
		me:u.ID,
		pageIdsCache:make(map[string]int),
		spotPageTemplate: spotPageTemplate,
		riverPageTemplate : riverPageTemplate,
		regionPageTemplate: regionPageTemplate,
		countryPageTemplate: countryPageTemplate,
		rootPageTemplate: rootPageTemplate,
		rateLimit:util.NewRateLimit(minDeltaBetweenRequests),
	}, nil
}

type HuskytmCatalogConnector struct {
	client              *wp.Client
	me                  int
	pageIdsCache        map[string]int
	spotPageTemplate    *template.Template
	riverPageTemplate   *template.Template
	regionPageTemplate  *template.Template
	countryPageTemplate *template.Template
	rootPageTemplate    *template.Template
	rateLimit           util.RateLimit
}

func (this *HuskytmCatalogConnector) Close() error {
	return nil
}

func (this *HuskytmCatalogConnector) CreateEmptyPageIfNotExistsAndReturnId(parent int, pageId int, title string) (int, string, bool, error) {
	log.Infof("Check page for id=%d", pageId)
	if pageId <= 0 {
		id, link, err := this.createPage(parent, title)
		created := err == nil
		return id, link, created, err
	}
	this.rateLimit.WaitIfNecessary()
	p, r, _, err := this.client.Pages().Get(pageId, emptyMap())
	if r.StatusCode == http.StatusNotFound || p.Status == "trash" {
		id, link, err := this.createPage(parent, title)
		created := err == nil
		return id, link, created, err
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
	return this.writePage(page.Id, this.spotPageTemplate, page.Spot.Title, page)
}
func (this *HuskytmCatalogConnector) WriteRiverPage(page common.RiverPageDto) error {
	return this.writePage(page.Id, this.riverPageTemplate, page.River.Title, page)
}
func (this *HuskytmCatalogConnector) WriteRegionPage(page common.RegionPageDto) error {
	return this.writePage(page.Id, this.regionPageTemplate, page.Region.Title, page)
}
func (this *HuskytmCatalogConnector) WriteCountryPage(page common.CountryPageDto) error {
	return this.writePage(page.Id, this.countryPageTemplate, page.Country.Title, page)
}

func (this *HuskytmCatalogConnector) WriteRootPage(page common.RootPageDto) error {
	return this.writePage(page.Id, this.rootPageTemplate, "Каталог водных препятствий", page)
}

func (this *HuskytmCatalogConnector) writePage(pageId int, tmpl *template.Template, title string, data interface{}) error {
	log.Infof("Write page %d for %s", pageId, title)
	htmlBuf := bytes.Buffer{}
	err := tmpl.Execute(&htmlBuf, data)
	if err != nil {
		log.Errorf("Can not process template", err)
		return err
	}
	page := wp.Page{
		Title:  wp.Title{Raw:title},
		Author:        this.me,
		Content:wp.Content{Raw:htmlBuf.String()},
	}

	this.rateLimit.WaitIfNecessary()
	_, r, b, err := this.client.Pages().Update(pageId, &page)
	if err != nil {
		log.Errorf("Connection failed. Code: %d Body: %s", r.StatusCode, string(b))
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

