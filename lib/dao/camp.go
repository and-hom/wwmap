package dao

import (
	"database/sql"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/pkg/errors"
)

func NewCampPostgresDao(postgresStorage PostgresStorage) CampDao {
	return &campStorage{
		PostgresStorage: postgresStorage,
		listQuery:       queries.SqlQuery("camp", "list"),
		findQuery:       queries.SqlQuery("camp", "find"),
		insertQuery:     queries.SqlQuery("camp", "insert"),
		updateQuery:     queries.SqlQuery("camp", "update"),
		removeQuery:     queries.SqlQuery("camp", "remove"),
	}
}

type campStorage struct {
	PostgresStorage
	listQuery   string
	findQuery   string
	insertQuery string
	updateQuery string
	removeQuery string
}

func (this campStorage) List() ([]Camp, error) {
	found, err := this.doFindList(this.listQuery, this.scan)
	if err != nil {
		return []Camp{}, err
	}
	return found.([]Camp), nil
}

func (this campStorage) Insert(camp Camp) (int64, error) {
	result, err := this.updateReturningId(this.insertQuery, this.campToArgs(false), true, camp)
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, errors.New("No id returned")
	}
	return result[0], nil
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

	err := rows.Scan(&camp.Id, &camp.Title, &camp.Description, &pointStr, &numTentPlaces)
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

		params := []interface{}{camp.Title, camp.Description, string(pathBytes), numTentPlaces}
		if withId {
			params = append([]interface{}{nullIf0(camp.Id)}, params...)
		}
		return params, nil
	}
}
