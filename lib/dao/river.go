package dao

import (
	"database/sql"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/lib/pq"
)

func NewRiverPostgresDao(postgresStorage PostgresStorage) RiverDao {
	return riverStorage{
		PostgresStorage:         postgresStorage,
		PropsManager:            PropertyManagerImpl{table: queries.SqlQuery("river", "table"), dao: &postgresStorage},
		findByTagsQuery:         queries.SqlQuery("river", "find-by-tags"),
		insideBoundsQuery:       queries.SqlQuery("river", "inside-bounds"),
		byIdQuery:               queries.SqlQuery("river", "by-id"),
		listAllQuery:            queries.SqlQuery("river", "all"),
		listByCountryQuery:      queries.SqlQuery("river", "by-country"),
		listByCountryFullQuery:  queries.SqlQuery("river", "by-country-full"),
		listByRegionQuery:       queries.SqlQuery("river", "by-region"),
		listByRegionFullQuery:   queries.SqlQuery("river", "by-region-full"),
		listByFirstLettersQuery: queries.SqlQuery("river", "by-first-letters"),
		insertQuery:             queries.SqlQuery("river", "insert"),
		updateFullQuery:         queries.SqlQuery("river", "update-full"),
		updateQuery:             queries.SqlQuery("river", "update"),
		deleteQuery:             queries.SqlQuery("river", "delete"),
		setVisibleQuery:         queries.SqlQuery("river", "set-visible"),
		findByTitlePartQuery:    queries.SqlQuery("river", "by-title-part"),
	}
}

type riverStorage struct {
	PostgresStorage
	PropsManager            PropertyManager
	findByTagsQuery         string
	insideBoundsQuery       string
	byIdQuery               string
	listAllQuery            string
	listByCountryQuery      string
	listByCountryFullQuery  string
	listByRegionQuery       string
	listByRegionFullQuery   string
	listByFirstLettersQuery string
	insertQuery             string
	updateQuery             string
	updateFullQuery         string
	deleteQuery             string
	setVisibleQuery         string
	findByTitlePartQuery    string
}

func (this riverStorage) FindTitles(titles []string) ([]RiverTitle, error) {
	return this.listRiverTitles(this.findByTagsQuery, pq.Array(titles))
}

func (this riverStorage) ListRiversWithBounds(bbox geo.Bbox, limit int, showUnpublished bool) ([]RiverTitle, error) {
	return this.listRiverTitles(this.insideBoundsQuery, bbox.Y1, bbox.X1, bbox.Y2, bbox.X2, limit, showUnpublished)
}

func (this riverStorage) FindByTitlePart(tPart string, limit, offset int) ([]RiverTitle, error) {
	return this.listRiverTitles(this.findByTitlePartQuery, tPart, tPart, limit, offset)
}

func (this riverStorage) Find(id int64) (River, error) {
	r, found, err := this.doFindAndReturn(this.byIdQuery, riverMapperFull, id)
	if err != nil {
		return River{}, err
	}
	if !found {
		return River{}, fmt.Errorf("River with id %d not found", id)
	}
	return r.(River), nil
}

func (this riverStorage) ListAll() ([]RiverTitle, error) {
	return this.listRiverTitles(this.listAllQuery)
}

func (this riverStorage) ListByCountry(countryId int64) ([]RiverTitle, error) {
	return this.listRiverTitles(this.listByCountryQuery, countryId)
}

func (this riverStorage) ListByCountryFull(countryId int64) ([]River, error) {
	return this.listRiverFull(this.listByCountryFullQuery, countryId)
}

func (this riverStorage) ListByRegion(regionId int64) ([]RiverTitle, error) {
	return this.listRiverTitles(this.listByRegionQuery, regionId)
}

func (this riverStorage) ListByRegionFull(regionId int64) ([]River, error) {
	return this.listRiverFull(this.listByRegionFullQuery, regionId)
}

func (this riverStorage) ListByFirstLetters(query string, limit int) ([]RiverTitle, error) {
	return this.listRiverTitles(this.listByFirstLettersQuery, query, limit)
}

func (this riverStorage) Insert(river River) (int64, error) {
	aliasesB, err := json.Marshal(river.Aliases)
	if err != nil {
		return 0, err
	}
	propsB, err := json.Marshal(river.Props)
	if err != nil {
		return 0, err
	}
	return this.insertReturningId(this.insertQuery, river.Region.Id, river.Title, string(aliasesB), river.Description, string(propsB))
}

func (this riverStorage) SaveFull(rivers ...River) error {
	vars := make([]interface{}, len(rivers))
	for i, p := range rivers {
		vars[i] = p
	}
	return this.performUpdates(this.updateFullQuery, func(entity interface{}) ([]interface{}, error) {
		_river := entity.(River)
		aliasesB, err := json.Marshal(_river.Aliases)
		if err != nil {
			return []interface{}{}, err
		}
		propsB, err := json.Marshal(_river.Props)
		if err != nil {
			return []interface{}{}, err
		}
		return []interface{}{_river.Id, _river.Region.Id, _river.Title, string(aliasesB), _river.Description, string(propsB)}, nil
	}, vars...)
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
		propsB, err := json.Marshal(_river.Props)
		if err != nil {
			return []interface{}{}, err
		}
		return []interface{}{_river.Id, _river.Region.Id, _river.Title, string(aliasesB), string(propsB)}, nil
	}, vars...)
}

func (this riverStorage) listRiverFull(query string, queryParams ...interface{}) ([]River, error) {
	found, err := this.doFindList(query, riverMapperFull, queryParams)
	if err != nil {
		return []River{}, err
	}
	return found.([]River), err
}

func (this riverStorage) listRiverTitles(query string, queryParams ...interface{}) ([]RiverTitle, error) {
	result, err := this.doFindList(query,
		func(rows *sql.Rows) (RiverTitle, error) {
			riverTitle := RiverTitle{}
			boundsStr := sql.NullString{}
			aliases := sql.NullString{}
			props := ""
			err := rows.Scan(&riverTitle.Id, &riverTitle.Region.Id, &riverTitle.Region.CountryId, &riverTitle.Title,
				&riverTitle.Region.Title, &riverTitle.Region.Fake, &boundsStr, &aliases, &props, &riverTitle.Visible)
			if err != nil {
				return RiverTitle{}, err
			}

			if boundsStr.Valid {
				riverTitle.Bounds, err = ParseBounds(boundsStr.String)
				if err != nil {
					log.Warnf("Can not parse rect or point %s for white water object %d: %v", boundsStr.String, riverTitle.Id, err)
				}
			}

			if aliases.Valid {
				err = json.Unmarshal([]byte(aliases.String), &riverTitle.Aliases)
			}
			if err != nil {
				return RiverTitle{}, err
			}
			err = json.Unmarshal([]byte(props), &riverTitle.Props)
			return riverTitle, err
		}, queryParams...)
	if err != nil {
		return []RiverTitle{}, err
	}
	return result.([]RiverTitle), nil
}

func ParseBounds(boundsStr string) (geo.Bbox, error) {
	var pgRect geo.PgPolygon
	err := json.Unmarshal([]byte(boundsStr), &pgRect)
	if err != nil {
		var pgPoint geo.GeoPoint
		err := json.Unmarshal([]byte(boundsStr), &pgPoint)
		if err != nil {
			return geo.Bbox{}, err
		}
		pgRect.Coordinates = point2rect(pgPoint)
	}

	return geo.Bbox{
		X1: pgRect.Coordinates[0][0].Lat,
		Y1: pgRect.Coordinates[0][0].Lon,
		X2: pgRect.Coordinates[0][2].Lat,
		Y2: pgRect.Coordinates[0][2].Lon,
	}, nil
}

func point2rect(pgPoint geo.GeoPoint) [][]geo.Point {
	// do not flip twice
	p := pgPoint.Coordinates
	return [][]geo.Point{{
		{
			Lat: p.Lat - 0.0001,
			Lon: p.Lon - 0.0001,
		},
		{
			Lat: p.Lat + 0.0001,
			Lon: p.Lon - 0.0001,
		},
		{
			Lat: p.Lat + 0.0001,
			Lon: p.Lon + 0.0001,
		},
		{
			Lat: p.Lat - 0.0001,
			Lon: p.Lon + 0.0001,
		},
	}}
}

func (this riverStorage) Remove(id int64, tx interface{}) error {
	log.Infof("Remove river %d", id)
	return this.performUpdatesWithinTxOptionally(tx, this.deleteQuery, IdMapper, id)
}

func (this riverStorage) Props() PropertyManager {
	return this.PropsManager
}

func (this riverStorage) SetVisible(id int64, visible bool) error {
	return this.performUpdates(this.setVisibleQuery, ArrayMapper, []interface{}{id, visible})
}

func riverMapperFull(rows *sql.Rows) (River, error) {
	river := River{}
	boundsStr := sql.NullString{}
	aliases := sql.NullString{}
	props := ""
	spotCounters := ""
	err := rows.Scan(&river.Id, &river.Region.Id, &river.Region.CountryId, &river.Title, &river.Region.Title, &river.Region.Fake, &boundsStr, &aliases, &river.Description, &river.Visible, &props, &spotCounters)
	if err != nil {
		return river, err
	}
	if aliases.Valid {
		err = json.Unmarshal([]byte(aliases.String), &river.Aliases)
	}
	if err != nil {
		return river, err
	}
	err = json.Unmarshal([]byte(props), &river.Props)
	if err != nil {
		return river, err
	}

	if boundsStr.Valid {
		river.Bounds, err = ParseBounds(boundsStr.String)
		if err != nil {
			log.Warnf("Can not parse rect or point %s for white water object %d: %v", boundsStr.String, river.Id, err)
		}
	}
	err = json.Unmarshal([]byte(spotCounters), &river.SpotCounters)
	return river, err
}
