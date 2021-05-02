package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	. "github.com/and-hom/wwmap/lib/dao"
	"time"
)

const REFERER_REMOVE_TTL time.Duration = 180 * 24 * time.Hour
const LEVEL_NULLS_REMOVE_TIME_OFFSET time.Duration = -14 * 24 * time.Hour

func main() {
	log.Infof("Starting wwmap")
	configuration := config.Load("")
	configuration.ConfigureLogger()

	storage := NewPostgresStorage(configuration.Db)
	refererDao := NewRefererPostgresDao(storage)
	err := refererDao.RemoveOlderThen(REFERER_REMOVE_TTL)
	if err != nil {
		log.Error(err)
	}

	levelDao := NewLevelPostgresDao(storage)
	err = levelDao.RemoveNullsBefore(JSONDate(time.Now().Add(LEVEL_NULLS_REMOVE_TIME_OFFSET)))
	if err != nil {
		log.Error(err)
	}
}
