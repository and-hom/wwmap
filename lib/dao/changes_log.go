package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"time"
)

func NewChangesLogPostgresDao(postgresStorage PostgresStorage) ChangesLogDao {
	return &changesLogStorage{
		PostgresStorage: postgresStorage,
		insertQuery:     queries.SqlQuery("changes-log", "insert"),
		listQuery:       queries.SqlQuery("changes-log", "list"),
		listAllQuery:    queries.SqlQuery("changes-log", "list-all"),
	}
}

type changesLogStorage struct {
	PostgresStorage
	insertQuery  string
	listQuery    string
	listAllQuery string
}

func (this changesLogStorage) Insert(entry ChangesLogEntry) error {
	_, err := this.updateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(ChangesLogEntry)
		return []interface{}{_e.ObjectType, _e.ObjectId, string(_e.AuthProvider),
			_e.ExtId, _e.Login, string(_e.Type), _e.Description, time.Time(_e.Time)}, nil
	}, true, entry)
	return err
}

func (this changesLogStorage) List(objectType string, objectId int64, limit int) ([]ChangesLogEntry, error) {
	return this.list(this.listQuery, objectType, objectId, limit)
}

func (this changesLogStorage) ListAll(limit int) ([]ChangesLogEntry, error) {
	return this.list(this.listAllQuery, limit)
}

func (this changesLogStorage) list(query string, args ...interface{}) ([]ChangesLogEntry, error) {
	lst, err := this.doFindList(query, func(rows *sql.Rows) (ChangesLogEntry, error) {
		result := ChangesLogEntry{}
		err := rows.Scan(&result.Id, &result.ObjectType, &result.ObjectId, &result.AuthProvider, &result.ExtId,
			&result.Login, &result.Type, &result.Description, &result.Time)
		return result, err
	}, args...)
	if err != nil {
		return []ChangesLogEntry{}, err
	}
	return lst.([]ChangesLogEntry), nil
}
