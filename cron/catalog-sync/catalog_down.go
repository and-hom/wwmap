package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/cron/catalog-sync/common"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm"
	"time"
)

func (this App) DoReadCatalog() {
	lastId, err := this.WwPassportDao.GetLastId(huskytm.SOURCE)
	if err != nil {
		this.Fatalf(err, "Can not connect get last ww passport entry id")
	}
	lastWwPassportIdStr := lastId.(time.Time).Format(huskytm.TIME_FORMAT)
	log.Infof("Get and store ww passport entries since %s", lastWwPassportIdStr)

	catalogConnector := this.getCachedCatalogConnector()

	wwPassportEntries, err := catalogConnector.PassportEntriesSince(lastWwPassportIdStr)
	if err != nil {
		log.Fatal("Can not get posts: ", err.Error())
	}

	log.Info("")
	for _, wwPassport := range wwPassportEntries {
		rivers, err := this.RiverDao.FindTitles([]string{wwPassport.River})
		if err != nil {
			log.Fatal("Can not find rivers for ww passport", wwPassport.Source, wwPassport.RemoteId, wwPassport.River)
		}
		switch len(rivers) {
		case 1:
			wwpts, err := this.WhiteWaterDao.ListByRiverAndTitle(rivers[0].Id, wwPassport.Title)
			if err != nil {
				log.Fatal("Can not find ww points for ww passport", wwPassport.Source, wwPassport.RemoteId, wwPassport.River, wwPassport.Title)
			}
			switch len(wwpts) {
			case 1:
				wwPassport.WwId = wwpts[0].Id
				err = this.WwPassportDao.Upsert(wwPassport)
				if err != nil {
					log.Fatal("Can not upsert ww passport", wwPassport, err)
				}
				this.findAndStoreImages(wwPassport, catalogConnector)
			case 0:
				log.Warn("No ww point found for %s - %s", wwPassport.River, wwPassport.Title)
				continue
			default:
				log.Warn("Can not explicitly detect ww point for ww passport", wwPassport.Source, wwPassport.RemoteId, wwPassport.River, wwPassport.River)
				continue
			}
		case 0:
			log.Warn("No river found", wwPassport.River)
			continue
		default:
			log.Warn("Can not explicitly detect river for ww passport", wwPassport.Source, wwPassport.RemoteId, wwPassport.River)
			continue
		}
	}
}

func (this App) findAndStoreImages(wwPassport dao.WWPassport, catalogConnector common.CatalogConnector) {
	log.Infof("Find images for ww passport %s-%s", wwPassport.Source, wwPassport.RemoteId)
	imgs, err := catalogConnector.GetImages(wwPassport.RemoteId)
	if err != nil {
		this.Fatalf(err, "Can not load images for ww passport %s-%s %d", wwPassport.Source, wwPassport.RemoteId, wwPassport.WwId)
	}
	_, err = this.ImgDao.Upsert(imgs...)
	if err != nil {
		this.Fatalf(err, "Can not upsert images for report %d", wwPassport.WwId)
	}
}
