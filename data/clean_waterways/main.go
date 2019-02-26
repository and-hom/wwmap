package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	. "github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
)

func main() {
	log.Infof("Starting wwmap")
	configuration := config.Load("")
	configuration.ChangeLogLevel()

	storage := NewPostgresStorage(configuration.Db)
	waterWayStorage := NewWaterWayPostgresDao(storage)

	err := waterWayStorage.ForEachWaterWay(func(ww WaterWay) (WaterWay, error) {
		nullPointCnt := 0
		for i := 0; i < len(ww.Path); i++ {
			if ww.Path[i].Lat == 0 && ww.Path[i].Lon == 0 {
				nullPointCnt++
			}
		}
		if nullPointCnt == 0 {
			return ww, nil
		}

		pointsNew := make([]geo.Point, len(ww.Path)-nullPointCnt)
		idx := 0
		for i := 0; i < len(ww.Path); i++ {
			if !(ww.Path[i].Lat == 0 && ww.Path[i].Lon == 0) {
				pointsNew[idx] = ww.Path[i]
				idx++
			}
		}
		log.Infof("%d points of %d removed\n", nullPointCnt, len(ww.Path))

		return WaterWay{
			Id:      ww.Id,
			OsmId:   ww.OsmId,
			Path:    pointsNew,
			Comment: ww.Comment,
			Type:    ww.Type,
			Title:   ww.Title,
			RiverId: ww.RiverId,
		}, nil
	}, "waterway_tmp")
	if err != nil {
		log.Fatal("Can not process: ", err)
	}
}
