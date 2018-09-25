package main

import (
	"github.com/and-hom/wwmap/lib/dao"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"fmt"
	"time"
	"strings"
)

const MAX_ATTACHED_IMGS = 300
const MISSING_IMAGE = "https://wwmap.ru/img/no-photo.png"
const PAGE_ID_PROP_NAME = "huskytm_page_id"

func filterRegions(regions []dao.Region) []dao.Region {
	result := make([]dao.Region, 0, len(regions))
	for _, r := range regions {
		if r.Title != "-" {
			result = append(result, r)
		}
	}
	return result
}

func filterCountries(countries []dao.Country) []dao.Country {
	result := make([]dao.Country, 0, len(countries))
	for _, c := range countries {
		if c.Title != "-" {
			result = append(result, c)
		}
	}
	return result
}

type DummyHasProperties struct {
	pageId int
}

func (this DummyHasProperties)Props() dao.PropertyManager {
	return DummyPropertyManager{pageId:this.pageId}
}

type DummyPropertyManager struct {
	pageId int
}

func (this DummyPropertyManager) GetIntProperty(name string, id int64) (int, error) {
	return this.pageId, nil
}
func (this DummyPropertyManager) SetIntProperty(name string, id int64, value int) error {
	return nil
}
func (this DummyPropertyManager) GetBoolProperty(name string, id int64) (bool, error) {
	return false, nil
}
func (this DummyPropertyManager) SetBoolProperty(name string, id int64, value bool) error {
	return nil
}
func (this *App) DoWriteCatalog() {
	for _, rpf := range this.catalogConnectors {
		err := rpf.Do(func(cc common.CatalogConnector) error {
			return this.doWriteCatalog(&cc)
		})
		if err != nil {
			log.Errorf("Can not access to source: %v", err)
		}
	}
}

func (this *App) doWriteCatalog(catalogConnector *common.CatalogConnector) error {
	fakeRegion := dao.Region{Id:0, Title:"-"}
	log.Info("Create missing ww passports")

	countries, err := this.CountryDao.List()
	if err != nil {
		log.Error("Can not list countries")
		return err
	}
	_, rootPageLink, err := this.createBlankPageIfNotExists(catalogConnector, DummyHasProperties{pageId:this.Configuration.RootPageId}, 0, "", 0)
	if err != nil {
		log.Error("Can not create blank root page if not exists")
		return err
	}
	countries = filterCountries(countries)
	countryLinks := []common.CountryLink{}
	for _, country := range countries {
		log.Infof("Upload country %s", country.Title)
		regions, err := this.RegionDao.List(country.Id)
		if err != nil {
			log.Errorf("Can not list regions for country %d", country.Id)
			return err
		}
		regions = filterRegions(regions)

		countryRivers, err := this.RiverDao.ListByCountryFull(country.Id)
		if err != nil {
			log.Errorf("Can not list rivers for country %d", country.Id)
			return err
		}

		if len(countryRivers) == 0 && len(regions) == 0 {
			log.Infof("Skip country %s - no rivers or regions", country.Title)
			continue
		}

		countryPageId, countryPageLink, err := this.createBlankPageIfNotExists(catalogConnector, this.CountryDao, country.Id, country.Title, this.Configuration.RootPageId)
		if err != nil {
			log.Error("Can not create blank country page if not exists")
			return err
		}

		countryRegionLinks := []common.LinkOnPage{}
		countryRiverLinks := []common.LinkOnPage{}
		for _, region := range regions {
			log.Infof("Upload region %s/%s", country.Title, region.Title)
			regionRivers, err := this.RiverDao.ListByRegionFull(region.Id)
			if err != nil {
				log.Errorf("Can not list rivers for region %d", region.Id)
				return err
			}
			if len(regionRivers) == 0 {
				log.Infof("Skip region %s - no rivers", region.Title)
				continue
			}

			regionPageId, regionPageLink, err := this.createBlankPageIfNotExists(catalogConnector, this.RegionDao, region.Id, region.Title, countryPageId)
			if err != nil {
				log.Error("Can not create blank root region if not exists")
				return err
			}

			riverLinks := []common.LinkOnPage{}
			for _, river := range regionRivers {
				riverPageLink := ""
				if river.SpotCounters.Ordered == river.SpotCounters.Total && river.SpotCounters.Total > 0 {
					riverPageLink, err = this.uploadRiver(catalogConnector, country, region, river, rootPageLink, countryPageLink, regionPageLink, regionPageId)
				}

				exportOk := err == nil && riverPageLink != ""
				log.Infof("Mark as exported: %v", exportOk)
				err2 := this.RiverDao.Props().SetBoolProperty("export_" + (*catalogConnector).SourceId(), river.Id, exportOk)
				if err != nil {
					return err
				}
				if err2 != nil {
					log.Errorf("Can not mark river %d as exported: %v", river.Id, err2)
				}
				if riverPageLink != "" {
					riverLinks = append(riverLinks, common.LinkOnPage{Title:river.Title, Url:riverPageLink})
				}
			}
			err = (*catalogConnector).WriteRegionPage(common.RegionPageDto{
				Id: regionPageId,
				Region:region,
				Country:country,
				Links:riverLinks,
				RootPageLink:rootPageLink,
				CountryPageLink:countryPageLink,
			})
			if err != nil {
				log.Errorf("Can not write region page %d", region.Id)
				return err
			}
			countryRegionLinks = append(countryRegionLinks, common.LinkOnPage{Title:region.Title, Url:regionPageLink})
		}
		for _, river := range countryRivers {
			log.Infof("Upload river %s/%s", country.Title, river.Title)
			riverPageLink := ""
			if river.SpotCounters.Ordered == river.SpotCounters.Total && river.SpotCounters.Total > 0 {
				riverPageLink, err = this.uploadRiver(catalogConnector, country, fakeRegion, river, rootPageLink, countryPageLink, "", countryPageId)
			}
			exportOk := err == nil && riverPageLink != ""
			log.Infof("Mark as exported: %v", exportOk)
			err2 := this.RiverDao.Props().SetBoolProperty("export_" + (*catalogConnector).SourceId(), river.Id, exportOk)
			if err != nil {
				return err
			}
			if err2 != nil {
				log.Errorf("Can not mark river %d as exported: %v", river.Id, err2)
			}
			if riverPageLink != "" {
				countryRiverLinks = append(countryRiverLinks, common.LinkOnPage{Title:river.Title, Url:riverPageLink})
			}
		}

		err = (*catalogConnector).WriteCountryPage(common.CountryPageDto{
			Id:countryPageId,
			Country:country,
			RegionLinks:countryRegionLinks,
			RiverLinks:countryRiverLinks,
			RootPageLink:rootPageLink,
		})
		if err != nil {
			log.Errorf("Can not write country page %d", country.Id)
			return err
		}

		countryLinks = append(countryLinks, common.CountryLink{
			LinkOnPage:  common.LinkOnPage{Title:country.Title, Url:countryPageLink},
			Code: country.Code,
		})
	}

	err = (*catalogConnector).WriteRootPage(common.RootPageDto{
		Id: this.Configuration.RootPageId,
		Links:countryLinks,
	})

	if err != nil {
		log.Errorf("Can not write root page contents")
	}
	return err
}

func (this *App) uploadRiver(catalogConnector *common.CatalogConnector, country dao.Country, region dao.Region, river dao.River,
rootPageLink, countryPageLink, regionPageLink string, parentPageId int) (string, error) {
	log.Infof("Upload river %s/%s/%s", country.Title, region.Title, river.Title)
	riverPageId, riverPageLink, spotLinks, needUpdate, err := this.writeSpots(catalogConnector, parentPageId, river, region, country, rootPageLink, countryPageLink, regionPageLink)
	if err != nil {
		log.Error("Can not create blank river page if not exists")
		return "", err
	}
	if !needUpdate {
		log.Infof("Skip river %s - no spots", river.Title)
		return "", nil
	}
	reports, err := this.reports(river.Id)
	if err != nil {
		log.Errorf("Can not get reports for river %d", river.Id)
		return "", err
	}
	err = (*catalogConnector).WriteRiverPage(common.RiverPageDto{
		Id:riverPageId,
		River:river,
		Region:region,
		Country:country,
		Links: spotLinks,
		RootPageLink:rootPageLink,
		CountryPageLink:countryPageLink,
		RegionPageLink:regionPageLink,
		MainImage:noImage(0),
		Reports:reports,
	})
	if err != nil {
		log.Errorf("Can not write river page %d", river.Id)
		return "", err
	}
	return riverPageLink, nil
}

const MAX_REPORTS_PER_SOURCE = 15

func (this *App) reports(riverId int64) ([]common.VoyageReportLink, error) {
	r, err := this.VoyageReportDao.List(riverId, MAX_REPORTS_PER_SOURCE)
	if err != nil {
		log.Errorf("Can not read report links: %v", err)
		return []common.VoyageReportLink{}, err
	}
	result := make([]common.VoyageReportLink, len(r))
	for i := 0; i < len(r); i++ {
		result[i] = common.VoyageReportLink{
			LinkOnPage:common.LinkOnPage{Title:r[i].Title, Url:r[i].Url},
			SourceLogo:this.ResourceBase + "/img/report_sources/" + strings.ToLower(r[i].Source) + ".png",
		}
	}
	return result, nil
}

func (this *App) createBlankPageIfNotExists(catalogConnector *common.CatalogConnector, dao dao.HasProperties, id int64, title string, parentId int) (int, string, error) {
	pageId, err := dao.Props().GetIntProperty(PAGE_ID_PROP_NAME, id)
	if err != nil {
		log.Errorf("Can not get page id for entity %d:%s", id, title)
		return 0, "", err
	}
	childPageId, link, created, err := (*catalogConnector).CreateEmptyPageIfNotExistsAndReturnId(id, parentId, pageId, title)
	if err != nil {
		log.Errorf("Can not create page for entity %d:%s", id, title)
		return 0, "", err
	}
	if created {
		log.Infof("Created page id=%d for %d - %s", childPageId, id, title)
		err := dao.Props().SetIntProperty(PAGE_ID_PROP_NAME, id, childPageId)
		if err != nil {
			log.Errorf("Can not set page id for entity %d:%s", id, title)
			return 0, "", err
		}
	}
	return childPageId, link, nil
}

func (this *App) writeSpots(catalogConnector *common.CatalogConnector, parentPageId int, river dao.River, region dao.Region, country dao.Country, rootPageLink, countryPageLink, regionPageLink string) (int, string, []common.SpotLink, bool, error) {
	spots, err := this.WhiteWaterDao.ListByRiverFull(river.Id)
	if err != nil {
		log.Errorf("Can not list spots for river %d", river.Id)
		return 0, "", []common.SpotLink{}, false, err
	}
	if len(spots) == 0 {
		return 0, "", []common.SpotLink{}, false, nil
	}

	riverPageId, riverPageLink, err := this.createBlankPageIfNotExists(catalogConnector, this.RiverDao, river.Id, river.Title, parentPageId)
	if err != nil {
		return 0, "", []common.SpotLink{}, false, err
	}

	spotLinks := []common.SpotLink{}
	for _, spot := range spots {
		log.Infof("Upload spot %s/%s", river.Title, spot.Title)
		spotPageId, spotPageLink, err := this.createBlankPageIfNotExists(catalogConnector, this.WhiteWaterDao, spot.Id, spot.Title, riverPageId)
		if err != nil {
			log.Errorf("Can not get attached images for %d", spot.Id)
			return 0, "", []common.SpotLink{}, false, err
		}
		imgs, err := this.ImgDao.List(spot.Id, MAX_ATTACHED_IMGS, dao.IMAGE_TYPE_IMAGE, true)
		if err != nil {
			log.Errorf("Can not get attached images for %d", spot.Id)
			return 0, "", []common.SpotLink{}, false, err
		}
		mainImg, err := this.mainImage(spot, imgs)
		if err != nil {
			log.Errorf("Can not get main image for spot %d", spot.Id)
			return 0, "", []common.SpotLink{}, false, err
		}
		err = (*catalogConnector).WriteSpotPage(common.SpotPageDto{
			Id:spotPageId,
			Spot:spot,
			River:river,
			Region:region,
			Country:country,
			MainImage:mainImg,
			Imgs:imgs,

			RootPageLink:rootPageLink,
			CountryPageLink:countryPageLink,
			RegionPageLink:regionPageLink,
			RiverPageLink:riverPageLink,
		})
		if err != nil {
			log.Errorf("Can not write spot page %d", spot.Id)
			return 0, "", []common.SpotLink{}, false, err
		}
		spotLinks = append(spotLinks, common.SpotLink{
			LinkOnPage: common.LinkOnPage{
				Title:spot.Title,
				Url:spotPageLink,
			},
			Category:common.CategoryStr(spot),
		})
	}
	return riverPageId, riverPageLink, spotLinks, true, nil
}

func (this *App) mainImage(spot dao.WhiteWaterPointFull, imgs []dao.Img) (dao.Img, error) {
	mainImg, found, err := this.ImgDao.GetMainForSpot(spot.Id)
	if err != nil {
		log.Errorf("Can not get main image for %d", spot.Id)
		return dao.Img{}, err
	}
	if found {
		this.processForWeb(&mainImg)
		return mainImg, nil
	}
	if len(imgs) > 0 {
		return imgs[0], nil
	} else {
		return noImage(spot.Id), nil
	}

}

func (this *App) processForWeb(img *dao.Img) {
	if img.Source == dao.IMG_SOURCE_WWMAP {
		img.Url = fmt.Sprintf(this.ImgUrlBase, img.Id)
		img.PreviewUrl = fmt.Sprintf(this.ImgUrlPreviewBase, img.Id)
	}
}
func noImage(spotId int64) dao.Img {
	return dao.Img{
		Id:0,
		Source:dao.IMG_SOURCE_WWMAP,
		Type:dao.IMAGE_TYPE_IMAGE,
		MainImage:true,
		Enabled:true,
		WwId:spotId,
		DatePublished:time.Now(),
		PreviewUrl: MISSING_IMAGE,
		Url: MISSING_IMAGE,
	}
}