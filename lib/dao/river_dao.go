package dao

import (
	log "github.com/Sirupsen/logrus"
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/geo"
	"fmt"
	"github.com/lib/pq"
)

type RiverStorage struct {
	PostgresStorage
}

func (this RiverStorage) FindTitles(titles []string) ([]RiverTitle, error) {
	return this.listRiverTitles("SELECT id, title,NULL FROM river WHERE title ilike ANY($1)", pq.Array(titles))
}

func (this RiverStorage) NearestRivers(point geo.Point, limit int) ([]RiverTitle, error) {
	pointBytes, err := json.Marshal(geo.NewGeoPoint(point))
	if err != nil {
		return []RiverTitle{}, err
	}
	fmt.Println(string(pointBytes))
	return this.listRiverTitles("SELECT id, title, NULL, aliases FROM (" +
		"SELECT ROW_NUMBER() OVER (PARTITION BY id ORDER BY distance ASC) AS r_num, id, title, distance, aliases FROM (" +
		"SELECT river.id AS id, river.title AS title, river.aliases AS aliases," +
		"ST_Distance(path,  ST_GeomFromGeoJSON($1)) AS distance FROM river INNER JOIN waterway ON river.id=waterway.river_id) ssq" +
		")sq WHERE r_num<=1 ORDER BY distance ASC LIMIT $2;", string(pointBytes), limit)
}

func (this RiverStorage) ListRiversWithBounds(bbox geo.Bbox, limit int) ([]RiverTitle, error) {
	return this.listRiverTitles("SELECT river.id, river.title, ST_AsGeoJSON(ST_Extent(white_water_rapid.point)), river.aliases FROM " +
		"river INNER JOIN white_water_rapid ON river.id=white_water_rapid.river_id " +
		"WHERE exists(SELECT 1 FROM white_water_rapid WHERE white_water_rapid.river_id=river.id and point && ST_MakeEnvelope($1,$2,$3,$4)) " +
		"GROUP BY river.id, river.title ORDER BY popularity DESC LIMIT $5", bbox.X1, bbox.Y1, bbox.X2, bbox.Y2, limit)
}

func (this RiverStorage) RiverById(id int64) (RiverTitle, error) {
	found, err := this.listRiverTitles("SELECT id,title,NULL,river.aliases AS aliases FROM river WHERE id=$1", id)
	if err != nil {
		return RiverTitle{}, err
	}
	if len(found) == 0 {
		return RiverTitle{}, fmt.Errorf("River with id %d not found", id)
	}
	return found[0], nil
}

func (this RiverStorage) listRiverTitles(query string, queryParams ...interface{}) ([]RiverTitle, error) {
	fmt.Println(query)
	result, err := this.doFindList(query,
		func(rows *sql.Rows) (RiverTitle, error) {
			riverTitle := RiverTitle{}
			boundsStr := sql.NullString{}
			aliases := ""
			err := rows.Scan(&riverTitle.Id, &riverTitle.Title, &boundsStr, &aliases)
			if err != nil {
				return RiverTitle{}, err
			}

			var pgRect PgPolygon
			if boundsStr.Valid {
				err = json.Unmarshal([]byte(boundsStr.String), &pgRect)
				if err != nil {
					var pgPoint PgPoint
					err = json.Unmarshal([]byte(boundsStr.String), &pgPoint)
					if err != nil {
						log.Warnf("Can not parse rect or point %s for white water object %d: %v", boundsStr.String, riverTitle.Id, err)
					}
					pgRect.Coordinates = [][]geo.Point{[]geo.Point{
						{
							Lat: pgPoint.Coordinates.Lat - 0.0001,
							Lon: pgPoint.Coordinates.Lon - 0.0001,
						},
						{
							Lat: pgPoint.Coordinates.Lat + 0.0001,
							Lon: pgPoint.Coordinates.Lon - 0.0001,
						},
						{
							Lat: pgPoint.Coordinates.Lat + 0.0001,
							Lon: pgPoint.Coordinates.Lon + 0.0001,
						},
						{
							Lat: pgPoint.Coordinates.Lat - 0.0001,
							Lon: pgPoint.Coordinates.Lon + 0.0001,
						},
					}, }
				}

				riverTitle.Bounds = geo.Bbox{
					X1:pgRect.Coordinates[0][0].Lon,
					Y1:pgRect.Coordinates[0][0].Lat,
					X2:pgRect.Coordinates[0][2].Lon,
					Y2:pgRect.Coordinates[0][2].Lat,
				}
			}

			err = json.Unmarshal([]byte(aliases), &riverTitle.Aliases)
			return riverTitle, err
		}, queryParams...)
	if (err != nil ) {
		return []RiverTitle{}, err
	}
	return result.([]RiverTitle), nil
}