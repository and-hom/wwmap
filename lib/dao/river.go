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
		PropsManager:PropertyManagerImpl{table:queries.SqlQuery("river", "table"), dao:&postgresStorage},
		findByTagsQuery: queries.SqlQuery("river", "find-by-tags"),
		nearestQuery: queries.SqlQuery("river", "nearest"),
		insideBoundsQuery: queries.SqlQuery("river", "inside-bounds"),
		byIdQuery:queries.SqlQuery("river", "by-id"),
		listByCountryQuery:queries.SqlQuery("river", "by-country"),
		listByRegionQuery:queries.SqlQuery("river", "by-region"),
		listByFirstLettersQuery:queries.SqlQuery("river", "by-first-letters"),
		insertQuery:queries.SqlQuery("river", "insert"),
		updateQuery:queries.SqlQuery("river", "update"),
		fixLinkedWaterWaysQuery:queries.SqlQuery("river", "fix-linked-waterways"),
		deleteLinkedWwptsQuery:queries.SqlQuery("river", "delete-linked-wwpts"),
		deleteLinkedReportsQuery:queries.SqlQuery("river", "delete-linked-reports"),
		deleteQuery:queries.SqlQuery("river", "delete"),
	}
}

type riverStorage struct {
	PostgresStorage
	PropsManager     PropertyManager
	findByTagsQuery          string
	nearestQuery             string
	insideBoundsQuery        string
	byIdQuery                string
	listByCountryQuery       string
	listByRegionQuery        string
	listByFirstLettersQuery  string
	insertQuery              string
	updateQuery              string
	fixLinkedWaterWaysQuery  string
	deleteLinkedWwptsQuery   string
	deleteLinkedReportsQuery string
	deleteQuery              string
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

func (this riverStorage) ListByFirstLetters(query string, limit int) ([]RiverTitle, error) {
	return this.listRiverTitles(this.listByFirstLettersQuery, query, limit)
}

func (this riverStorage) Insert(river RiverTitle) (int64, error) {
	aliasesB, err := json.Marshal(river.Aliases)
	if err != nil {
		return 0, err
	}
	return this.insertReturningId(this.insertQuery, river.RegionId, river.Title, string(aliasesB))
}

func (this riverStorage) Save(rivers ...RiverTitle) error {
	vars := make([]interface{}, len(rivers))
	for i, p := range rivers {
		vars[i] = p
	}
	return this.performUpdates(this.updateQuery, func(entity interface{}) ([]interface{}, error) {
		_river := entity.(RiverTitle)
		aliasesB, err := json.Marshal(_river.Aliases)
		if err != nil {
			return []interface{}{}, err
		}
		return []interface{}{_river.Id, _river.RegionId, _river.Title, string(aliasesB)}, nil
	}, vars...)
}

func (this riverStorage) listRiverTitles(query string, queryParams ...interface{}) ([]RiverTitle, error) {
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

func (this riverStorage) Remove(id int64) error {
	log.Infof("Remove river %d", id)
	tx, err := this.Begin()
	if err != nil {
		return err
	}
	defer tx.Close();

	for _, q := range []string{this.fixLinkedWaterWaysQuery, this.deleteLinkedWwptsQuery, this.deleteLinkedReportsQuery, this.deleteQuery} {
		err = tx.performUpdates(q, idMapper, id)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (this riverStorage) Props() PropertyManager {
	return this.PropsManager
}
