package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
)

func main() {
	log.Infof("Starting wwmap river tracks bind")
	configuration := config.Load("")
	configuration.ConfigureLogger()
	storage := dao.NewPostgresStorage(configuration.Db)

	waterWayDao := dao.NewWaterWayPostgresDao(storage)

	log.Info("Simplify waterways preserving ref points")
	batchSize := 10000

	persister, err := waterWayDao.PathSimplifiedPersister()
	if err != nil {
		log.Fatal(err)
	}

	modifiedCnt := 0
	totalCount := 0
	for {
		log.Infof("Select %d rows", batchSize)
		found, err := waterWayDao.List(batchSize, totalCount)
		if err != nil {
			log.Fatal("Can't select waterways ", err)
		}
		if len(found) == 0 {
			log.Infof("%d waterways successfully processed. Complete!", totalCount)
			break
		}
		totalCount += len(found)
		log.Infof("%d: Take %d waterways", totalCount, len(found))

		for i := 0; i < len(found); i++ {
			if fixWays(&(found[i])) {
				modifiedCnt++
				if err := persister.Add(found[i].Id, found[i].PathSimplified); err != nil {
					log.Warnf("Can't fix path for %d: %v", found[i].Id, err)
				}
			}
		}

		log.Infof("%d rows processed", totalCount)

		if persister.Close() != nil {
			log.Fatal(err)
		}
		persister, err = waterWayDao.PathSimplifiedPersister()
		if err != nil {
			log.Fatal(err)
		}
	}

	if persister.Close() != nil {
		log.Fatal(err)
	}

	log.Infof("%d of %d waterways modified", modifiedCnt, totalCount)
}

func fixWays(waterway *dao.WaterWay4PathCorrection) bool {
	sPos := 0
	insertedCnt := 0
	changed := false
	newPathSimplified := make([]geo.Point, 0, len(waterway.Path))

	for pos := 0; pos < len(waterway.Path); pos++ {
		if sPos < len(waterway.PathSimplified) && waterway.Path[pos] == waterway.PathSimplified[sPos] {
			// Point is in both path and path simplified
			newPathSimplified = append(newPathSimplified, waterway.PathSimplified[sPos])
			sPos++
			continue
		}

		for refPtIdx := 0; refPtIdx < len(waterway.CrossPoints); refPtIdx++ {
			if waterway.CrossPoints[refPtIdx] == waterway.Path[pos] {
				// Point is cross point and not in path simplified. Add.
				changed = true

				newPathSimplified = append(newPathSimplified, waterway.Path[pos])
				//log.Debugf("Insert point %v for %d", waterway.Path[pos], waterway.Id)
				insertedCnt++
				break
			}
		}

		// Point is not in path simplified and is not cross point - skip
	}

	if sPos < len(waterway.PathSimplified)-1 {
		log.Warnf("Tail of simplified path is not contained in original path for way id=%d", waterway.Id)
	}

	if changed {
		log.Infof("Path %d has %d modifications", waterway.Id, insertedCnt)
		waterway.PathSimplified = newPathSimplified
	}

	return changed
}
