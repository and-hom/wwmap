package main

import (
	"github.com/and-hom/wwmap/cron/river-height-mapper/altitue_map"
	"github.com/and-hom/wwmap/cron/river-height-mapper/util"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	log "github.com/sirupsen/logrus"
	"math"
)

func main() {
	log.Infof("Starting wwmap river height mapper")
	configuration := config.Load("")
	configuration.ConfigureLogger()
	storage := dao.NewPostgresStorage(configuration.Db)

	waterWayDao := dao.NewWaterWayPostgresDao(storage)
	srtmDao := dao.NewSrtmPostgresDao(storage)

	log.Info("Compute river height vectors")

	persister, err := waterWayDao.PathHeightPersister()
	if err != nil {
		log.Fatal(err)
	}

	found, err := waterWayDao.ListWithRiver()
	if err != nil {
		log.Fatal("Can't select waterways ", err)
	}

	log.Infof("Selected %d rows", len(found))

	worldAltitudeMap := altitue_map.WorldAltitudeMap(srtmDao)

	okCount := 0

	for i := 0; i < len(found); i++ {
		path := found[i].Path
		log.Infof("Process waterway %d", found[i].Id)
		heights := make([]int, len(path))
		dists := make([]float64, len(path))
		ok := false
		tgAlpha := math.MaxFloat64

		for idx, point := range path {
			if idx == 0 {
				dists[idx] = 0
			} else {
				dists[idx] = point.DistanceTo(path[idx-1])
				tgAlpha = (point.Lat - path[idx-1].Lat) / (point.Lon - path[idx-1].Lon)
			}

			latSec, lonSec := int(point.Lat*3600), int(point.Lon*3600)

			var h int
			if tgAlpha == math.MaxFloat64 || tgAlpha == math.Inf(1) || tgAlpha == math.Inf(-1) {
				h, err = getMinHeightAround(worldAltitudeMap, latSec, lonSec, 2)
			} else {
				h, err = getMinHeightAcrossRiverValley(worldAltitudeMap, latSec, lonSec, tgAlpha, 8)
			}
			if err != nil {
				switch e := err.(type) {
				case *altitue_map.RasterNotFound:
					if e.FirstTime() {
						log.Errorf("Can't get point %d %d from raster for %s: %s", latSec, lonSec, point.String(), err)
					}
				default:
					log.Errorf("Can't get point %d %d from raster for %s: %s", latSec, lonSec, point.String(), err)
				}
				continue
			}

			heights[idx] = h

			ok = true
		}
		if ok {
			log.Infof("Height map found!")
			util.RemoveHeightGrowing(heights)

			if err := persister.Add(found[i].Id, heights, dists); err != nil {
				log.Warnf("Can't fix path for %d: %v", found[i].Id, err)
			} else {
				okCount++
			}
		}
	}

	log.Infof("%d rows updated", okCount)

	if persister.Close() != nil {
		log.Fatal(err)
	}

	log.Infof(worldAltitudeMap.GetMissingRasters())
}

func getMinHeightAcrossRiverValley(
	raster altitue_map.AltitudeMap,
	latSec int,
	lonSec int,
	tgAlpha float64,
	step int,
) (int, error) {
	var result = math.MaxInt32
	var lastErr error

	for i := -step; i < step; i++ {
		dLat, dLon := GetVectorNormPoint(i, tgAlpha)
		pointLat := latSec - int(dLat)
		pointLon := lonSec + int(dLon)

		c, err := raster.Get(pointLat, pointLon)
		if err != nil {
			if dLat==0 && dLon==0 {
				// Не надо прерывать поиск, если не найдена одна из точек сбоку - только если не найдена
				// центральная точка
				lastErr = err
			}
			continue
		}
		if c < result {
			result = c
		}
	}

	return result, lastErr
}

func GetVectorNormPoint(step int, tgAlpha float64) (float64, float64) {
	if tgAlpha == 0.0 {
		return -1.0, 0.0
	}
	if math.IsInf(tgAlpha, 1) {
		return 0.0, 1.0
	}
	if math.IsInf(tgAlpha, -1) {
		return 0.0, -1.0
	}
	dLon := math.Sqrt(tgAlpha * tgAlpha / (1 + tgAlpha*tgAlpha))
	dLat := -dLon / tgAlpha
	return float64(step) * dLat, float64(step) * dLon
}

func getMinHeightAround(altitudeMap altitue_map.AltitudeMap, latSec int, lonSec int, step int) (int, error) {
	var result = math.MaxInt32
	var lastErr error
	for i := -step; i <= step; i++ {
		for j := -step; j <= step; j++ {
			c, err := altitudeMap.Get(latSec+i, lonSec+j)
			if err != nil {
				lastErr = err
				continue
			}
			if c < result {
				result = c
			}
		}
	}
	return result, lastErr
}

