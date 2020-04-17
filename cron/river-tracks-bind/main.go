package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
)

func main() {
	log.Infof("Starting wwmap river tracks bind")
	configuration := config.Load("")
	configuration.ConfigureLogger()
	storage := dao.NewPostgresStorage(configuration.Db)

	waterWayDao := dao.NewWaterWayPostgresDao(storage)

	log.Info("Bind tracks to rivers")
	err := waterWayDao.BindWaterwaysToRivers()
	if err != nil {
		log.Fatalf("Can't bind tracks to river: ", err)
	}
}
