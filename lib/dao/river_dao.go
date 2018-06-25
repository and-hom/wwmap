package dao

import (
	log "github.com/Sirupsen/logrus"
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/geo"
	"fmt"
	"github.com/lib/pq"
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewRiverPostgresDao(postgresStorage PostgresStorage) RiverDao {
	return riverStorage{
		PostgresStorage: postgresStorage,
		findByTagsQuery: queries.SqlQuery("river", "find-by-tags"),
		nearestQuery: queries.SqlQuery("river", "nearest"),
		insideBoundsQuery: queries.SqlQuery("river", "inside-bounds"),
		byIdQuery:queries.SqlQuery("river", "by-id"),
		listByCountryQuery:queries.SqlQuery("river", "by-country"),
		listByRegionQuery:queries.SqlQuery("river", "by-region"),
		updateQuery:queries.SqlQuery("river", "update"),
	}
}

type riverStorage struct {
	PostgresStorage
	findByTagsQuery    string
	nearestQuery       string
	insideBoundsQuery  string
	byIdQuery          string
	listByCountryQuery string
	listByRegionQuery  string
	updateQuery  string
}

func (this riverStorage) FindTitles(titles []string) ([]RiverTitle, error) {
	return this.listRiverTitles(this.findByTagsQuery, pq.Array(titles))
}

func (this riverStorage) NearestRivers(point geo.Point, limit int) ([]RiverTitle, error) {
	pointBytes, err := json.Marshal(geo.NewGeoPoint(point))
	if err != nil {
		return []RiverTitle{}, err
	}
	return this.listRiverTitles(this.nearestQuery, string(pointBytes), limit)
}

func (this riverStorage) ListRiversWithBounds(bbox geo.Bbox, limit int) ([]RiverTitle, error) {
	return this.listRiverTitles(this.insideBoundsQuery, bbox.X1, bbox.Y1, bbox.X2, bbox.Y2, limit)
}

func (this riverStorage) Find(id int64) (RiverTitle, error) {
	found, err := this.listRiverTitles(this.byIdQuery, id)
	if err != nil {
		return RiverTitle{}, err
	}
	if len(found) == 0 {
		return RiverTitle{}, fmt.Errorf("River with id %d not found", id)
	}
	return found[0], nil
}

func (this riverStorage) ListByCountry(countryId int64) ([]RiverTitle, error) {
	return this.listRiverTitles(this.listByCountryQuery, countryId)
}

func (this riverStorage) ListByRegion(regionId int64) ([]RiverTitle, error) {
	return this.listRiverTitles(this.listByRegionQuery, regionId)
}

func (this riverStorage) Save(river RiverTitle) error {
	return this.performUpdates(this.updateQuery, func(entity interface{}) ([]interface{}, error){
		_river := entity.(RiverTitle)
		aliasesB, err := json.Marshal(_river.Aliases)
		if err!=nil {
			return []interface{}{}, err
		}
		return []interface{}{_river.RegionId, _river.Title, string(aliasesB), _river.Id}, nil
	}, river)
}

func (this riverStorage) listRiverTitles(query string, queryParams ...interface{}) ([]RiverTitle, error) {
	fmt.Println(query)
	result, err := this.doFindList(query,
		func(rows *sql.Rows) (RiverTitle, error) {
			riverTitle := RiverTitle{}
			boundsStr := sql.NullString{}
			aliases := sql.NullString{}
			err := rows.Scan(&riverTitle.Id, &riverTitle.RegionId, &riverTitle.Title, &boundsStr, &aliases)
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

			if aliases.Valid {
				err = json.Unmarshal([]byte(aliases.String), &riverTitle.Aliases)
			}
			return riverTitle, err
		}, queryParams...)
	if (err != nil ) {
		return []RiverTitle{}, err
	}
	return result.([]RiverTitle), nil
}