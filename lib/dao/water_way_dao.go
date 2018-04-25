package dao

import (
	"encoding/json"
	"github.com/and-hom/wwmap/lib/geo"
	"database/sql"
	log "github.com/Sirupsen/logrus"
)

type WaterWayStorage struct {
	PostgresStorage
}

func (this WaterWayStorage) AddWaterWays(waterways ...WaterWay) error {
	vars := make([]interface{}, len(waterways))
	for i, p := range waterways {
		vars[i] = p
	}
	return this.performUpdates("INSERT INTO waterway(osm_id, title, type, comment, path) VALUES ($1, $2, $3, $4, ST_GeomFromGeoJSON($5))",
		func(entity interface{}) ([]interface{}, error) {
			waterway := entity.(WaterWay)

			pathBytes, err := json.Marshal(geo.NewLineString(waterway.Path))
			if err != nil {
				return nil, err
			}
			return []interface{}{waterway.OsmId, waterway.Title, waterway.Type, waterway.Comment, string(pathBytes)}, nil;
		}, vars...)
}

func (this WaterWayStorage) UpdateWaterWay(waterway WaterWay) error {
	return this.performUpdates("UPDATE waterway SET path=ST_GeomFromGeoJSON($1) WHERE osm_id=$2",
		func(entity interface{}) ([]interface{}, error) {
			waterway := entity.(WaterWay)

			pathBytes, err := json.Marshal(geo.NewLineString(waterway.Path))
			if err != nil {
				return nil, err
			}
			return []interface{}{string(pathBytes), waterway.OsmId}, nil;
		}, waterway)
}

func (this WaterWayStorage) ForEachWaterWay(transformer func(WaterWay) (WaterWay, error), tmpTable string) error {
	rows, err := this.db.Query("SELECT id, osm_id, river_Id, title, type, comment, ST_AsGeoJSON(path) FROM waterway")
	if err != nil {
		return err
	}
	defer rows.Close()

	tx, err := this.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO " + tmpTable + "(id, osm_id, river_Id, title, type, comment, path) VALUES ($1, $2, $3, $4, $5, $6, ST_GeomFromGeoJSON($7))")
	if err != nil {
		return err
	}
	defer stmt.Close();

	i := 0
	for rows.Next() {
		waterWay := WaterWay{}
		osmId := sql.NullInt64{}
		riverId := sql.NullInt64{}
		pathStr := ""
		rows.Scan(&waterWay.Id, &osmId, &riverId, &waterWay.Title, &waterWay.Type, &waterWay.Comment, &pathStr)
		if osmId.Valid {
			waterWay.OsmId = osmId.Int64
		}
		if riverId.Valid {
			waterWay.RiverId = riverId.Int64
		}
		var path geo.LineString
		err = json.Unmarshal([]byte(pathStr), &path)
		if err != nil {
			log.Errorf("Can not parse path \"%s\": %v", path, err)
			return err
		}
		waterWay.Path = path.Coordinates

		waterWayNew, err := transformer(waterWay)
		if err != nil {
			log.Errorf("Can not transofrm waterway %d: %v", waterWay.Id, err)
			return err
		}

		pathBytesNew, err := json.Marshal(geo.NewLineString(waterWayNew.Path))
		if err != nil {
			log.Errorf("Can not serialize path %v: %v", waterWayNew.Path, err)
			return err
		}
		riverIdNew := sql.NullInt64{
			Valid:waterWayNew.RiverId > 0,
			Int64:waterWayNew.RiverId,
		}
		stmt.Exec(waterWayNew.Id, waterWayNew.OsmId, riverIdNew, waterWayNew.Title, waterWayNew.Type, waterWayNew.Comment, string(pathBytesNew))
		i++
		if i % 1000 == 0 {
			//	err := tx.Commit()
			//	if err != nil {
			//		log.Errorf("Can not commit: %v", err)
			//		return err
			//	}
			log.Info("Progress: ", i)
		}
	}
	tx.Commit()
	return nil
}