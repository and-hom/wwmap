package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
)

const BIND_OK_PROP = "waterways_bind"

func main() {
	log.Infof("Starting wwmap river tracks bind")
	configuration := config.Load("")
	configuration.ChangeLogLevel()
	storage := dao.NewPostgresStorage(configuration.Db)

	riverDao := dao.NewRiverPostgresDao(storage)
	waterWayDao := dao.NewWaterWayPostgresDao(storage)

	rivers, err := riverDao.ListAll()
	if err != nil {
		log.Fatal("Can't list rivers", err)
	}

	for _, river := range rivers {
		bindOk, found := river.Props[BIND_OK_PROP]
		bindOkBool, isBool := bindOk.(bool)
		if !found || isBool && !bindOkBool {

			ids, err := waterWayDao.BindToRiver(river.Id, river.TitleVariants())
			if err != nil {
				log.Error("Can't bind tracks to river: ", err)
				continue
			}
			log.Infof("%d waterways bind to river %s (%v) (%d): %v", len(ids), river.Title, river.Aliases, river.Id, ids)

			river.Props[BIND_OK_PROP] = true
			err = riverDao.Save(river)
			if err != nil {
				log.Error("Can't save river: ", err)
			}
		}
	}
}
