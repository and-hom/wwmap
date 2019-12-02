package dao

import (
	"database/sql"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/geo"
)

func NewCampPostgresDao(postgresStorage PostgresStorage) CampDao {
	return &campStorage{
		PostgresStorage:       postgresStorage,
		listQuery:             queries.SqlQuery("camp", "list"),
		findWithinBoundsQuery: queries.SqlQuery("camp", "find-witin-bounds"),
		findQuery:             queries.SqlQuery("camp", "find"),
		insertQuery:           queries.SqlQuery("camp", "insert"),
		updateQuery:           queries.SqlQuery("camp", "update"),
		removeQuery:           queries.SqlQuery("camp", "remove"),
	}
}

type campStorage struct {
	PostgresStorage
	listQuery             string
	findWithinBoundsQuery string
	findQuery             string
	insertQuery           string
	updateQuery           string
	removeQuery           string
}

func (this campStorage) List() ([]Camp, error) {
	found, err := this.doFindList(this.listQuery, this.scan)
	if err != nil {
		return []Camp{}, err
	}
	return found.([]Camp), nil
}

func (this campStorage) FindWithinBounds(bbox geo.Bbox) ([]Camp, error) {
	found, err := this.doFindList(this.findWithinBoundsQuery, this.scan, bbox.Y1, bbox.X1, bbox.Y2, bbox.X2)
	if err != nil {
		return []Camp{}, err
	}
	return found.([]Camp), nil
}

func (this campStorage) Insert(camp ...Camp) ([]int64, error) {
	data := make([]interface{}, len(camp))
	for i := 0; i < len(camp); i++ {
		data[i] = camp[i]
	}
	result, err := this.updateReturningId(this.insertQuery, this.campToArgs(false), true, data...)
	if err != nil {
		return []int64{}, err
	}
	return result, nil
}

func (this campStorage) Update(camp Camp) error {
	return this.performUpdates(this.removeQuery, this.campToArgs(true), camp)
}

func (this campStorage) Find(id int64) (Camp, bool, error) {
	camp, found, err := this.doFindAndReturn(this.findQuery, this.scan, id)
	if err != nil {
		return Camp{}, false, err
	}
	return camp.(Camp), found, err
}

func (this campStorage) Remove(id int64, tx interface{}) error {
	return this.performUpdatesWithinTxOptionally(tx, this.removeQuery, IdMapper, id)
}

func (this campStorage) scan(rows *sql.Rows) (Camp, error) {
	camp := Camp{}
	pointStr := ""
	numTentPlaces := sql.NullInt64{}
	osmId := sql.NullInt64{}

	err := rows.Scan(&camp.Id, &osmId, &camp.Title, &camp.Description, &pointStr, &numTentPlaces)
	if err != nil {
		return camp, err
	}

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
