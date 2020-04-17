package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/vodinfo-eye/graduation"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
)

func main() {
	log.Infof("Starting wwmap vodinfo sensor data processing")
	configuration := config.Load("")
	configuration.ConfigureLogger()

	graduator, err := graduation.NewPercentileGladiator(0.1, 0.1)
	if err != nil {
		log.Fatal(err)
	}

	postgres := dao.NewPostgresStorage(configuration.Db)

	levelSensorDao := dao.NewLevelSensorPostgresDao(postgres)
	levelDao := dao.NewLevelPostgresDao(postgres)

	sensors, err := levelSensorDao.List()
	if err != nil {
		log.Fatal(err)
	}
	for _, sensor := range sensors {
		graduation.ReCalculateSensorMinMax(graduator, levelSensorDao, levelDao, sensor.Id)
	}
}
