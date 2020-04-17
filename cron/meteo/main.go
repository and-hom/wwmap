package main

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
)

func main() {
	log.Infof("Starting wwmap meteo")
	configuration := config.Load("")
	configuration.ConfigureLogger()
	storage := dao.NewPostgresStorage(configuration.Db)

	meteoDao := dao.NewMeteoPostgresDao(storage)
	meteoPointDao := dao.NewMeteoPointPostgresDao(storage)

	api := NewYandexWeatherApi(configuration.MeteoToken)

	points, err := meteoPointDao.List()
	if err != nil {
		log.Fatal("Can not read meteo points: ", err)
	}

	for _, point := range points {
		if !point.CollectData {
			continue
		}

		meteos, err := api.Get(point.Point)
		if err != nil {
			pStr, _ := point.Point.MarshalJSON()
			log.Errorf("Can not read meteo data for %d %s %v from api: %v", point.Id, point.Title, string(pStr), err)
			continue
		}
		if len(meteos) == 0 {
			pStr, _ := point.Point.MarshalJSON()
			log.Errorf("Empty data for %d %s %v from api: %v", point.Id, point.Title, string(pStr), err)
			continue
		}
		for i := 0; i < len(meteos); i++ {
			meteos[i].PointId = point.Id
		}
		j, _ := json.Marshal(meteos[0])
		log.Debugf("Weather is %s for %d %s", string(j), point.Id, point.Title)

		err = meteoDao.Insert(meteos[0])
		if err != nil {
			log.Error("Can not save point meteo data: ", err)
			continue
		}
	}
}
