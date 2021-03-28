package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/config"
)

type App struct {
	WhiteWaterDao dao.WhiteWaterDao
	WaterWayDao   dao.WaterWayDao
}

func CreateApp() App {
	configuration := config.Load("")
	pgStorage := dao.NewPostgresStorage(configuration.Db)
	return App{
		WhiteWaterDao:dao.NewWhiteWaterPostgresDao(pgStorage),
		WaterWayDao:dao.NewWaterWayPostgresDao(pgStorage),
	}
}

func main() {
	log.Infof("Starting wwmap spot sorter")
	app := CreateApp()
	app.PerformReordering()
}

func (this App) PerformReordering() {
	// 0. Detect rivers changed since last execution
	// Select spots with automatic ordering flag grouped by river by criteria:
	//	River has more then one spot with automatic ordering flag and with different last_ordered timestamp.
	riverIds, err := this.WhiteWaterDao.AutoOrderingRiverIds()
	if err != nil {
		log.Fatalf("Can not list river ids for spot reordering: %v+", err)
	}
	log.Infof("%d rivers for reordering: %v", len(riverIds), riverIds)

	// 1. Select all gpx tracks from whaterways (OSM) for
	for _, riverId := range riverIds {
		log.Infof("Find tracks for river %d", riverId)
		waterways, err := this.WaterWayDao.DetectForRiver(riverId)
		if err != nil {
			log.Fatalf("Can not get nearest waterways for river %d: %v", riverId, err)
		}

		log.Infof("Found %d (%v) waterways for river %d", len(waterways), ids(waterways), riverId)
		orderIdx := this.getOrderIdx(riverId, waterways)
		err = this.WhiteWaterDao.UpdateOrderIdx(orderIdx)
		if err != nil {
			log.Fatalf("Can not write reordering info %v for river %d: %v", orderIdx, riverId, err)
		}
	}
}

func (this App) getOrderIdx(riverId int64, waterways []dao.WaterWay) map[int64]int {
	switch len(waterways) {
	case 0:
		log.Warn("Can not reorder spots - no waterways")
		return make(map[int64]int)
	case 1:
		orderIdx, err := this.WhiteWaterDao.DistanceFromBeginning(riverId, waterways[0].Path)
		if err != nil {
			log.Fatalf("Can not get order idx for river %d: %v", riverId, err)
		}
		return orderIdx
	default:
		log.Warn("More then one waterway: reordering is not implemented yet")
		return make(map[int64]int)
	}
}

func ids(waterways []dao.WaterWay) []int64 {
	result := make([]int64, len(waterways))
	for i := 0; i < len(waterways); i++ {
		result[i] = waterways[i].Id
	}
	return result
}