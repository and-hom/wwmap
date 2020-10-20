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

func NewCampPostgresDao(postgresStorage PostgresStorage) CampDao {
	return &campStorage{
		riverLinksStorage: riverLinksStorage{
			PostgresStorage:        postgresStorage,
			listRefsByRiverQuery:   queries.SqlQuery("camp", "list-refs-by-river"),
			insertRefsQuery:        queries.SqlQuery("camp", "insert-refs"),
			deleteRefsQuery:        queries.SqlQuery("camp", "delete-refs"),
			deleteRefsByRiverQuery: queries.SqlQuery("camp", "delete-refs-by-river"),
			listRivers:             queries.SqlQuery("linked-entity", "list-rivers"),
		},
		listQuery:             queries.SqlQuery("camp", "list"),
		findWithinBoundsQuery: queries.SqlQuery("camp", "find-witin-bounds"),
		findQuery:             queries.SqlQuery("camp", "find"),
		insertQuery:           queries.SqlQuery("camp", "insert"),
		updateQuery:           queries.SqlQuery("camp", "update"),
		removeQuery:           queries.SqlQuery("camp", "remove"),
	}
}

type campStorage struct {
	riverLinksStorage
	listQuery             string
	findWithinBoundsQuery string
	findQuery             string
	insertQuery           string
	updateQuery           string
	removeQuery           string
}

func (this campStorage) List(withRivers bool) ([]Camp, error) {
	found, err := this.DoFindList(this.listQuery, this.scan)
	if err != nil {
		return []Camp{}, err
	}

	camps := found.([]Camp)

	if withRivers {
		if err := this.enrichWithRiverData(convertCamps(&camps)); err != nil {
			return nil, err
		}
	}

	return camps, nil
}

func convertCamps(transfers *[]Camp) []ILinkedEntity {
	result := make([]ILinkedEntity, len(*transfers))
	for i := 0; i < len(*transfers); i++ {
		result[i] = &(*transfers)[i]
	}
	return result
}

func (this campStorage) FindWithinBounds(bbox geo.Bbox) ([]Camp, error) {
	found, err := this.DoFindList(this.findWithinBoundsQuery, this.scan, bbox.Y1, bbox.X1, bbox.Y2, bbox.X2)
	if err != nil {
		return []Camp{}, err
	}
	return found.([]Camp), nil
}

func (this campStorage) InsertMultiple(camp ...Camp) ([]int64, error) {
	data := make([]interface{}, len(camp))
	for i := 0; i < len(camp); i++ {
		data[i] = camp[i]
	}
	result, err := this.UpdateReturningId(this.insertQuery, this.campToArgs(false), true, data...)
	if err != nil {
		return []int64{}, err
	}
	return result, nil
}

func (this campStorage) Insert(camp Camp) (int64, error) {
	fields, err := this.campToArgs(false)(camp)
	if err != nil {
		return 0, err
	}
	return this.insert(this.insertQuery, camp.Rivers, fields)
}

func (this campStorage) Update(camp Camp) error {
	fields, err := this.campToArgs(true)(camp)
	if err != nil {
		return err
	}
	return this.update(this.updateQuery, camp.Id, camp.Rivers, fields)
}

func (this campStorage) Find(id int64) (Camp, bool, error) {
	camp, found, err := this.DoFindAndReturn(this.findQuery, this.scan, id)
	fmt.Println(this.findQuery, camp)
	if err != nil {
		return Camp{}, false, err
	}
	return camp.(Camp), found, err
}

func (this campStorage) Remove(id int64, tx interface{}) error {
	return this.PerformUpdatesWithinTxOptionally(tx, this.removeQuery, IdMapper, id)
}

func (this campStorage) scan(rows *sql.Rows) (Camp, error) {
	camp := Camp{}
	pointStr := ""
	numTentPlaces := sql.NullInt64{}
	osmId := sql.NullInt64{}
	rivers := pq.Int64Array{}

	err := rows.Scan(&camp.Id, &osmId, &camp.Title, &camp.Description, &pointStr, &numTentPlaces, &rivers)
	if err != nil {
		return camp, err
	}

	camp.Rivers = []int64(rivers)

	var pgPoint geo.GeoPoint
	err = json.Unmarshal([]byte(pointStr), &pgPoint)
	if err != nil {
		log.Errorf("Can not parse centroid point %s for camp %d: %v", pointStr, camp.Id, err)
		return camp, err
	}
	camp.Point = pgPoint.Coordinates.Flip()

	if numTentPlaces.Valid {
		camp.NumTentPlaces = uint16(numTentPlaces.Int64)
	} else {
		camp.NumTentPlaces = 0
	}

	if osmId.Valid {
		camp.OsmId = osmId.Int64
	} else {
		camp.OsmId = 0
	}
	return camp, nil
}

func (this campStorage) campToArgs(withId bool) func(entity interface{}) ([]interface{}, error) {
	return func(entity interface{}) ([]interface{}, error) {
		camp := entity.(Camp)
		pathBytes, err := json.Marshal(geo.NewPgGeoPoint(camp.Point))
		if err != nil {
			return nil, err
		}

		var numTentPlaces sql.NullInt64
		if camp.NumTentPlaces > 0 {
			numTentPlaces.Valid = true
			numTentPlaces.Int64 = int64(camp.NumTentPlaces)
		} else {
			numTentPlaces.Valid = false
		}

		var osmId sql.NullInt64
		if camp.OsmId > 0 {
			osmId.Valid = true
			osmId.Int64 = int64(camp.OsmId)
		} else {
			osmId.Valid = false
		}

		params := []interface{}{osmId, camp.Title, camp.Description, string(pathBytes), numTentPlaces}
		if withId {
			params = append([]interface{}{nullIf0(camp.Id)}, params...)
		}
		return params, nil
	}
}
