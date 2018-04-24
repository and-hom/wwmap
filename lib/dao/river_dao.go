package dao

import (
	log "github.com/Sirupsen/logrus"
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/geo"
	"fmt"
)

type RiverStorage struct {
	PostgresStorage
}

func (this RiverStorage) NearestRivers(point geo.Point, limit int) ([]RiverTitle, error) {
	pointBytes, err := json.Marshal(geo.NewGeoPoint(point))
	if err != nil {
		return []RiverTitle{}, err
	}
	return this.listRiverTitles("SELECT id, title,NULL FROM (" +
		"SELECT river.id AS id, river.title AS title, ST_Distance(path,  ST_GeomFromGeoJSON($1)) as distance " +
		"FROM river INNER JOIN waterway ON river.id=waterway.river_id" +
		") sq GROUP BY id, title ORDER BY min(distance) ASC LIMIT $2", string(pointBytes), limit)
}

func (this RiverStorage) ListRiversWithBounds(bbox geo.Bbox, limit int) ([]RiverTitle, error) {
	return this.listRiverTitles("SELECT river.id, river.title, ST_AsGeoJSON(ST_Extent(waterway.path)) FROM " +
		"river INNER JOIN waterway ON river.id=waterway.river_id " +
		"WHERE exists(SELECT 1 FROM white_water_rapid WHERE white_water_rapid.river_id=river.id) AND " +
		"path && ST_MakeEnvelope($1,$2,$3,$4) " +
		"GROUP BY river.id, river.title ORDER BY popularity DESC LIMIT $5", bbox.X1, bbox.Y1, bbox.X2, bbox.Y2, limit)
}

func (this RiverStorage) RiverById(id int64) (RiverTitle, error) {
	found, err := this.listRiverTitles("SELECT id,title FROM river WHERE id=$1", id)
	if err != nil {
		return RiverTitle{}, err
	}
	if len(found) == 0 {
		return RiverTitle{}, fmt.Errorf("River with id %d not found", id)
	}
	return found[0], nil
}

func (this RiverStorage) listRiverTitles(query string, queryParams ...interface{}) ([]RiverTitle, error) {
	result, err := this.doFindList(query,
		func(rows *sql.Rows) (RiverTitle, error) {
			id := int64(-1)
			title := ""
			boundsStr := sql.NullString{}
			err := rows.Scan(&id, &title, &boundsStr)
			if err != nil {
				return RiverTitle{}, err
			}

			var pgRect PgPolygon
			if boundsStr.Valid {
				err = json.Unmarshal([]byte(boundsStr.String), &pgRect)
				if err != nil {
					log.Errorf("Can not parse rect %s for white water object %d: %v", boundsStr.String, id, err)
				}
			}
			bounds := geo.Bbox{
				X1:pgRect.Coordinates[0][0].Lon,
				Y1:pgRect.Coordinates[0][0].Lat,
				X2:pgRect.Coordinates[0][2].Lon,
				Y2:pgRect.Coordinates[0][2].Lat,
			}

			return RiverTitle{
				Id:id,
				Title:title,
				Bounds: bounds,
			}, nil
		}, queryParams...)
	if (err != nil ) {
		return []RiverTitle{}, err
	}
	return result.([]RiverTitle), nil
}