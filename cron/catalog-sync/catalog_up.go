package main

import (
	"github.com/and-hom/wwmap/lib/dao"
	"strconv"
	"github.com/and-hom/wwmap/cron/catalog-sync/huskytm"
	log "github.com/Sirupsen/logrus"
)

const MAX_ATTACHED_IMGS = 300

func (this App) DoWriteCatalog() {
	log.Info("Create missing ww passports")
	catalogConnector := this.getCachedCatalogConnector()

	wwpts, err := this.WhiteWaterDao.ListWithPath()
	if err != nil {
		this.Fatalf(err, "Can not connect list ww points")
	}

	for i := 0; i < len(wwpts); i++ {
		point := &wwpts[i]
		log.Debug("Write point ", point.Title)
		exists, err := catalogConnector.Exists(point.Path)
		if err != nil {
			this.Fatalf(err, "Can not check page exists")
		}
		if !exists {
			log.Debug("Not exists - create")
			imgs, err := this.ImgDao.List(point.Id, 1)
			if err != nil {
				this.Fatalf(err, "Can not get images for ww point %d", point.Id)
			}
			imgId := -1
			if len(imgs) > 0 {
				imgId64, err := strconv.ParseInt(imgs[0].RemoteId, 10, 32)
				if err != nil {
					this.Fatalf(err, "Can not parse img remote id", imgs[0].RemoteId)
				}
				imgId = int(imgId64)
			}
			this.createCatalogEntry(point, imgId)
		} else {
			log.Debugf("Point %s exists", point.Path)
		}
	}
}

func (this App) createCatalogEntry(p *dao.WhiteWaterPointWithPath, imgId int) {
	parent_id, err := this.mkdirRecursiveQuietly(parent(p.Path))
	if err != nil {
		this.Fatalf(err, "Can not create directories")
	}
	catalogConnector := this.getCachedCatalogConnector()
	imgs, err := this.ImgDao.List(p.Id, MAX_ATTACHED_IMGS)
	if err != nil {
		this.Fatalf(err, "Can not get attached images for %d", p.Id)
	}
	err = catalogConnector.Create(p.WhiteWaterPoint, parent_id, imgs)
	if err != nil {
		this.Fatalf(err, "Can not create passport")
	}
}

func parent(path []string) []string {
	if len(path) == 0 {
		log.Fatal("Can not create parent of empty path")
	}
	return path[:len(path) - 1]
}

func (this *App) mkdirRecursiveQuietly(p []string) (int, error) {
	if len(p) == 0 {
		return huskytm.ROOT_PAGE, nil
	}
	catalogConnector := this.getCachedCatalogConnector()
	pp := parent(p)

	parent, err := this.mkdirRecursiveQuietly(pp)
	if err != nil {
		return 0, err
	}

	exists, err := catalogConnector.Exists(p)
	if err != nil {
		return 0, err
	}
	title := p[len(p) - 1]
	if !exists {
		log.Debug("Mk dir ", p)
		return catalogConnector.CreatePage(title, parent)
	} else {
		log.Debug("Dir exists ", p)
		return catalogConnector.GetId(title, parent)
	}
}
