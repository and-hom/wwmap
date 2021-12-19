package altitue_map

import (
	"encoding/json"
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	log "github.com/sirupsen/logrus"
)

type AltitudeMap interface {
	Get(latSec int, lonSec int) (int, error)
}

type AltitudeMapWithDiagnostics interface {
	AltitudeMap
	GetMissingRasters() string
}

func WorldAltitudeMap(srtmDao dao.SrtmDao) AltitudeMapWithDiagnostics {
	return &worldAltitudeMap{
		rasters:        make(map[string]AltitudeMap),
		missingRasters: make(map[string]bool),
		srtmDao:        srtmDao,
	}
}

type worldAltitudeMap struct {
	rasters        map[string]AltitudeMap
	missingRasters map[string]bool
	srtmDao        dao.SrtmDao
}

func (this worldAltitudeMap) Get(latSec int, lonSec int) (int, error) {
	latDegrees, lonDegrees := latSec/3600, lonSec/3600
	raster, err := this.getRaster(latDegrees, lonDegrees)
	if err != nil {
		return 0, err
	}
	return raster.Get(latSec%3600, lonSec%3600)
}

func (this *worldAltitudeMap) getRaster(latDegrees int, lonDegrees int) (AltitudeMap, error) {
	key := this.key(latDegrees, lonDegrees)

	if _, missing := this.missingRasters[key]; missing {
		return nil, &RasterNotFound{lat: latDegrees, lon: lonDegrees, firstTime: false}
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
			return nil, &RasterNotFound{lat: latDegrees, lon: lonDegrees, firstTime: true}
		}

		log.Infof("Raster found for lat=%d lon=%d", latDegrees, lonDegrees)
		this.rasters[key] = raster
	}

	return raster, err
}

func (this *worldAltitudeMap) key(latDegrees, lonDegrees int) string {
	return fmt.Sprintf("%d %d", latDegrees, lonDegrees)
}

func (this *worldAltitudeMap) GetMissingRasters() string {
	b, err := json.MarshalIndent(this.missingRasters, "", "  ")
	if err != nil {
		log.Error(err)
	}
	return string(b)
}
