package huskytm

import (
	"github.com/and-hom/wwmap/lib/dao"
	wp "github.com/and-hom/go-wordpress"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"github.com/and-hom/wwmap/cron/catalog-sync/bindata"
	"fmt"
	"net/http"
	"html/template"
	log "github.com/Sirupsen/logrus"
	"bytes"
)

const ROOT_PAGE = 1739

func GetCatalogConnector(login, password string) (common.CatalogConnector, error) {
	client := wp.NewClient(&wp.Options{
		BaseAPIURL: API_BASE, // example: `http://192.168.99.100:32777/wp-json/wp/v2`
		Username:   login,
		Password:   password,
	})
	u, r, b, e := client.Users().Me(emptyMap())
	if e != nil {
		return nil, fmt.Errorf("Connection failed: %s", e.Error())
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Connection failed. Code: %d Body: %s", r.StatusCode, string(b))
	}
	spotPageTemplate, e := template.New("spot").Parse(string(bindata.MustAsset("spot-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile template: %s", e.Error())
	}
	return &HuskytmCatalogConnector{
		client:client,
		me:u.ID,
		pageIdsCache:make(map[string]int),
		spotPageTemplate: spotPageTemplate,
	}, nil
}

type HuskytmCatalogConnector struct {
	client           *wp.Client
	me               int
	pageIdsCache     map[string]int
	spotPageTemplate *template.Template
}

func (this *HuskytmCatalogConnector) Close() error {
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

func (this *HuskytmCatalogConnector) Exists(key []string) (bool, error) {
	log.Debug("Check exists ", key)
	parentId := ROOT_PAGE
	for i := 0; i < len(key); i++ {
		id, err := this.GetId(key[i], parentId)
		if err != nil {
			switch err.(type) {
			default:
				return false, err
			case PageNotFoundError:
				log.Debug("Not exists ", key[i])
				return false, nil
			}
		}
		log.Debug("Exists ", key[i])
		parentId = id
	}
	return true, nil
}

func (this *HuskytmCatalogConnector) Create(wwPoint dao.WhiteWaterPointFull, parent int, _ []dao.Img) error {
	htmlBuf := bytes.Buffer{}
	err := this.spotPageTemplate.Execute(&htmlBuf, map[string]interface{}{
		"spot": wwPoint,
	})
	if err!=nil {
		log.Errorf("Can not process template", err)
		return err
	}
	page := wp.Page{
		Title:  wp.Title{Raw:wwPoint.Title},
		Author:        this.me,
		Parent:        parent,
		Status:        "publish",
		Content:wp.Content{Raw:htmlBuf.String()},
	}
	fmt.Printf(page.Content.Raw)

	_, r, b, err := this.client.Pages().Create(&page)
	if err != nil {
		log.Errorf("Connection failed. Code: %d Body: %s", r.StatusCode, string(b))
		return err
	}

	return nil
}

func (this *HuskytmCatalogConnector) CreatePage(title string, parent int) (int, error) {
	p, r, b, err := this.client.Pages().Create(&wp.Page{
		Title:        wp.Title{Raw:title},
		Author:        this.me,
		Parent:        parent,
		Status:        "publish",
	})
	if err != nil {
		log.Errorf("Connection failed. Code: %d Body: %s", r.StatusCode, string(b))
		return 0, err
	}
	return p.ID, nil
}

func (this *HuskytmCatalogConnector) GetId(title string, parent int) (int, error) {
	cacheId := fmt.Sprintf("%d-%s", parent, title)
	idFromCache, isInCache := this.pageIdsCache[cacheId]
	if isInCache {
		log.Infof("Get id for %s from cache: %d", title, idFromCache)
		return idFromCache, nil
	} else {
		log.Infof("Do real request: get id for %s child of %d", title, parent)
	}

	params := emptyMap()
	if parent > 0 {
		params["parent"] = parent
	}
	found, err := paginate(func(p interface{}) ([]interface{}, *http.Response, []byte, error) {
		f, r, b, e := this.client.Pages().List(params)
		res := make([]interface{}, len(f))
		for i := 0; i < len(f); i++ {
			res[i] = f[i]
		}
		return res, r, b, e
	}, params)

	if err != nil {
		return 0, err
	}
	for _, p := range found {
		//log.Debugf("Search by name for %s: %s", title, p.(wp.Page).Title.Rendered)
		if p.(wp.Page).Title.Rendered == title {
			id := p.(wp.Page).ID
			this.pageIdsCache[cacheId] = id
			return id, nil
		}
	}
	return 0, PageNotFoundError{fmt.Sprintf("Can not find page with name \"%s\" as child of %d", title, parent)}
}

type PageNotFoundError struct {
	msg string
}

func (this PageNotFoundError) Error() string {
	return this.msg
}

