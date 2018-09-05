package main

import (
	"github.com/and-hom/wwmap/lib/dao"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"fmt"
	"time"
)

const MAX_ATTACHED_IMGS = 300
const MISSING_IMAGE = "https://wwmap.ru/editor/img/no-photo.png"
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
func (this *App) DoWriteCatalog() {
	fakeRegion := dao.Region{Id:0, Title:"-"}
	log.Info("Create missing ww passports")
	catalogConnector := this.getCachedCatalogConnector()
	countries, err := this.CountryDao.List()
	if err != nil {
		this.Fatalf(err, "Can not list countries")
	}
	_, rootPageLink := this.createBlankPageIfNotExists(DummyHasProperties{pageId:this.Configuration.RootPageId}, 0, "", 0)
	countries = filterCountries(countries)
	countryLinks := []common.LinkOnPage{}
	for _, country := range countries {
		log.Infof("Upload country %s", country.Title)
		regions, err := this.RegionDao.List(country.Id)
		if err != nil {
			this.Fatalf(err, "Can not list regions for country %d", country.Id)
		}
		regions = filterRegions(regions)

		rivers, err := this.RiverDao.ListByCountry(country.Id)
		if err != nil {
			this.Fatalf(err, "Can not list rivers for country %d", country.Id)
		}

		if len(rivers) == 0 && len(regions) == 0 {
			log.Infof("Skip country %s - no rivers or regions", country.Title)
			continue
		}

		countryPageId, countryPageLink := this.createBlankPageIfNotExists(this.CountryDao, country.Id, country.Title, this.Configuration.RootPageId)

		countryRegionLinks := []common.LinkOnPage{}
		countryRiverLinks := []common.LinkOnPage{}
		for _, region := range regions {
			log.Infof("Upload region %s/%s", country.Title, region.Title)
			regionRivers, err := this.RiverDao.ListByRegion(region.Id)
			if err != nil {
				this.Fatalf(err, "Can not list rivers for region %d", region.Id)
			}
			if len(regionRivers) == 0 {
				log.Infof("Skip region %s - no rivers", region.Title)
				continue
			}

			regionPageId, regionPageLink := this.createBlankPageIfNotExists(this.RegionDao, region.Id, region.Title, countryPageId)

			riverLinks := []common.LinkOnPage{}
			for _, river := range regionRivers {
				log.Infof("Upload river %s/%s/%s", country.Title, region.Title, river.Title)
				riverPageId, riverPageLink, spotLinks, needUpdate := this.writeSpots(regionPageId, river, region, country, rootPageLink, countryPageLink, regionPageLink)
				if !needUpdate {
					log.Infof("Skip river %s - no spots", river.Title)
					continue
				}
				err := catalogConnector.WriteRiverPage(riverPageId, river, region, country, spotLinks, rootPageLink, countryPageLink, regionPageLink)
				if err != nil {
					this.Fatalf(err, "Can not write river page %d", river.Id)
				}
				riverLinks = append(riverLinks, common.LinkOnPage{Title:river.Title, Url:riverPageLink})
			}
			err = catalogConnector.WriteRegionPage(regionPageId, region, country, riverLinks, rootPageLink, countryPageLink)
			if err != nil {
				this.Fatalf(err, "Can not write region page %d", region.Id)
			}
			countryRegionLinks = append(countryRegionLinks, common.LinkOnPage{Title:region.Title, Url:regionPageLink})
		}
		for _, river := range rivers {
			log.Infof("Upload river %s/%s", country.Title, river.Title)
			riverPageId, riverPageLink, spotLinks, needUpdate := this.writeSpots(countryPageId, river, fakeRegion, country, rootPageLink, countryPageLink, "")
			if !needUpdate {
				log.Infof("Skip river %s - no spots", river.Title)
				continue
			}
			err := catalogConnector.WriteRiverPage(riverPageId, river, fakeRegion, country, spotLinks, rootPageLink, countryPageLink, "")
			if err != nil {
				this.Fatalf(err, "Can not write river page %d", river.Id)
			}
			countryRiverLinks = append(countryRiverLinks, common.LinkOnPage{Title:river.Title, Url:riverPageLink})
		}

		err = catalogConnector.WriteCountryPage(countryPageId, country, countryRegionLinks, countryRiverLinks, rootPageLink)
		if err != nil {
			this.Fatalf(err, "Can not write country page %d", country.Id)
		}

		countryLinks = append(countryLinks, common.LinkOnPage{Title:country.Title, Url:countryPageLink})
	}

	catalogConnector.WriteRootPage(this.Configuration.RootPageId, countryLinks)
}

func (this *App) createBlankPageIfNotExists(dao dao.HasProperties, id int64, title string, parentId int) (int, string) {
	catalogConnector := this.getCachedCatalogConnector()
	pageId, err := dao.Props().GetIntProperty(PAGE_ID_PROP_NAME, id)
	if err != nil {
		this.Fatalf(err, "Can not get page id for entity %d:%s", id, title)
	}
	childPageId, link, created, err := catalogConnector.CreateEmptyPageIfNotExistsAndReturnId(parentId, pageId, title)
	if err != nil {
		this.Fatalf(err, "Can not create page for entity %d:%s", id, title)
	}
	if created {
		log.Infof("Created page id=%d for %d - %s", childPageId, id, title)
		err := dao.Props().SetIntProperty(PAGE_ID_PROP_NAME, id, childPageId)
		if err != nil {
			this.Fatalf(err, "Can not set page id for entity %d:%s", id, title)
		}
	}
	return childPageId, link
}

func (this *App) writeSpots(parentPageId int, river dao.RiverTitle, region dao.Region, country dao.Country, rootPageLink, countryPageLink, regionPageLink string) (int, string, []common.LinkOnPage, bool) {
	spots, err := this.WhiteWaterDao.ListByRiverFull(river.Id)
	if err != nil {
		this.Fatalf(err, "Can not list spots for river %d", river.Id)
	}
	if len(spots) == 0 {
		return 0, "", []common.LinkOnPage{}, false
	}

	catalogConnector := this.getCachedCatalogConnector()
	riverPageId, riverPageLink := this.createBlankPageIfNotExists(this.RiverDao, river.Id, river.Title, parentPageId)

	spotLinks := []common.LinkOnPage{}
	for _, spot := range spots {
		log.Infof("Upload spot %s/%s", river.Title, spot.Title)
		spotPageId, spotPageLink := this.createBlankPageIfNotExists(this.WhiteWaterDao, spot.Id, spot.Title, riverPageId)
		imgs, err := this.ImgDao.List(spot.Id, MAX_ATTACHED_IMGS, dao.IMAGE_TYPE_IMAGE, true)
		if err != nil {
			this.Fatalf(err, "Can not get attached images for %d", spot.Id)
		}
		err = catalogConnector.WriteSpotPage(spotPageId, spot, river, region, country, this.mainImage(spot, imgs), imgs,
			rootPageLink, countryPageLink, regionPageLink, riverPageLink)
		if err != nil {
			this.Fatalf(err, "Can not write spot page %d", spot.Id)
		}
		spotLinks = append(spotLinks, common.LinkOnPage{Title:spot.Title, Url:spotPageLink})
	}
	return riverPageId, riverPageLink, spotLinks, true
}

func (this *App) mainImage(spot dao.WhiteWaterPointFull, imgs []dao.Img) dao.Img {
	mainImg, found, err := this.ImgDao.GetMainForSpot(spot.Id)
	if err != nil {
		this.Fatalf(err, "Can not get main image for %d", spot.Id)
	}
	if found {
		this.processForWeb(&mainImg)
		return mainImg
	}
	if len(imgs) > 0 {
		return imgs[0]
	} else {
		return dao.Img{
			Id:0,
			Source:dao.IMG_SOURCE_WWMAP,
			Type:dao.IMAGE_TYPE_IMAGE,
			MainImage:true,
			Enabled:true,
			WwId:spot.Id,
			DatePublished:time.Now(),
			PreviewUrl: MISSING_IMAGE,
			Url: MISSING_IMAGE,
		}
	}

}

func (this *App) processForWeb(img *dao.Img) {
	if img.Source == dao.IMG_SOURCE_WWMAP {
		img.Url = fmt.Sprintf(this.ImgUrlBase, img.Id)
		img.PreviewUrl = fmt.Sprintf(this.ImgUrlPreviewBase, img.Id)
	}
}