package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"time"
)

func NewWWPassportPostgresDao(postgresStorage PostgresStorage) WwPassportDao {
	return wwPassportStorage{
		PostgresStorage: postgresStorage,
		getLastIdQuery:  queries.SqlQuery("ww-passport", "get-last-id"),
	}
}

type wwPassportStorage struct {
	PostgresStorage
	getLastIdQuery string
}

func (this wwPassportStorage) Upsert(wwPassport ...WWPassport) error {
	return nil
}

func (this wwPassportStorage) GetLastId(source string) (interface{}, error) {
	lastDate, found, err := this.DoFindAndReturn(this.getLastIdQuery, func(rows *sql.Rows) (time.Time, error) {
		lastDate := time.Unix(0, 0)
		err := rows.Scan(&lastDate)
		return lastDate, err
	}, source)
	if err != nil {
		return nil, err
	}
	if !found {
		return time.Unix(0, 0), nil
	}
	return lastDate, nil
}
