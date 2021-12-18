package main

import (
	"encoding/json"
	"fmt"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
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
	calc := heightCalculator{}
	calc.init(srtmDao)

	okCount := 0

	for i := 0; i < len(found); i++ {
		path := found[i].Path
		log.Infof("Process waterway %d", found[i].Id)
		heights := make([]int32, len(path))
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

			latSec, lonSec := getAngleSecondFrac(point.Lat), getAngleSecondFrac(point.Lon)
			raster, err := calc.getRaster(point)
			if err != nil {
				continue
			}

			var h int32
			if tgAlpha == math.MaxFloat64 {
				h, err = getMinHeightAround(raster, latSec, lonSec, 2)
			} else {
				h, err = getMinHeightAcrossRiverValley(raster, latSec, lonSec, tgAlpha, 8)
			}
			if err != nil {
				log.Errorf("Can't get point %d %d from raster for %s: %s", latSec, lonSec, point.String(), err)
				continue
			}

			heights[idx] = h

			ok = true
		}
		if ok {
			log.Infof("Height map found!")
			RemoveHeightGrowing(heights)

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

	calc.printMissingRasters()
}

func getMinHeightAcrossRiverValley(
	raster geo.Bytearea2D,
	latSec int,
	lonSec int,
	tgAlpha float64,
	step int,
) (int32, error) {
	var result int32 = math.MaxInt32
	var lastErr error

	for i := -step; i < step; i++ {
		dLat, dLon := GetVectorNormPoint(i, tgAlpha)
		x := 3600 - latSec + int(dLat)
		y := lonSec + int(dLon)

		// TODO: тут проверяем выход за границы растра
		// Надо сделать получение из соседнего растра при необходимости
		if i != 0 && (x < 0 || y < 0 || x > 3600 || y > 3600) {
			continue
		}

		c, err := raster.Get(x, y)
		if err != nil {
			lastErr = err
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

func getMinHeightAround(raster geo.Bytearea2D, latSec int, lonSec int, step int) (int32, error) {
	var result int32 = math.MaxInt32
	var lastErr error
	for i := -step; i <= step; i++ {
		for j := -step; j <= step; j++ {
			c, err := raster.Get(3600-latSec+i, lonSec+j)
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

type heightCalculator struct {
	rasters        map[string]geo.Bytearea2D
	missingRasters map[string]bool
	srtmDao        dao.SrtmDao
}

func (this *heightCalculator) init(srtmDao dao.SrtmDao) {
	this.rasters = make(map[string]geo.Bytearea2D)
	this.missingRasters = make(map[string]bool)
	this.srtmDao = srtmDao
}

func (this *heightCalculator) key(latDegrees, lonDegrees int) string {
	return fmt.Sprintf("%d %d", latDegrees, lonDegrees)
}

func (this *heightCalculator) getRaster(point geo.Point) (geo.Bytearea2D, error) {
	latDegrees, lonDegrees := int(point.Lat), int(point.Lon)
	key := this.key(latDegrees, lonDegrees)

	if _, missing := this.missingRasters[key]; missing {
		return nil, fmt.Errorf("Raster not found for %d %d", latDegrees, lonDegrees)
	}

	raster, found := this.rasters[key]
	var err error

	if !found {
		log.Infof("Getting raster for %d %d", latDegrees, lonDegrees)
		raster, found, err = this.srtmDao.GetRaster(latDegrees, lonDegrees)
		if err != nil {
			log.Errorf("Failed to get raster for lat=%d lon=%d: %s", latDegrees, lonDegrees, err)
			this.missingRasters[key] = true
			return nil, err
		}
		if !found {
			this.missingRasters[key] = true
			log.Warnf("Not found raster for lat=%d lon=%d", latDegrees, lonDegrees)
			return nil, fmt.Errorf("Raster not found for %d %d", latDegrees, lonDegrees)
		}

		log.Infof("Raster found for lat=%d lon=%d", latDegrees, lonDegrees)
		this.rasters[key] = raster
	}

	return raster, err
}

func (this *heightCalculator) printMissingRasters() {
	b, err := json.MarshalIndent(this.missingRasters, "", "  ")
	if err != nil {
		log.Error(err)
	}
	log.Infof("Missing %s", string(b))
}

func getAngleSecondFrac(value float64) int {
	_, frac := math.Modf(value)
	return int(frac * 3600)
}
