package main

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/config"
	"time"
)

const REMOVE_TTL time.Duration = 180 * 24 * time.Hour

func main() {
	log.Infof("Starting wwmap")
	configuration := config.Load("")
	configuration.ChangeLogLevel()

	storage := NewPostgresStorage(configuration.Db)
	refererDao := NewRefererPostgresDao(storage)
	err := refererDao.RemoveOlderThen(REMOVE_TTL)
	if err!=nil {
		log.Fatal(err)
	}
}
