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
	storage := NewPostgresStorage(configuration.DbConnString)
	refererDao := NewRefererPostgresDao(storage)
	refererDao.RemoveOlderThen(REMOVE_TTL)

}
