package dao

import (
	"time"
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewWWPassportPostgresDao(postgresStorage PostgresStorage) WwPassportDao {
	return wwPassportStorage{
		PostgresStorage: postgresStorage,
		getLastIdQuery: queries.SqlQuery("ww-passport", "get-last-id"),
	}
}

type wwPassportStorage struct {
	PostgresStorage
	getLastIdQuery string
}

func (this wwPassportStorage) Upsert(wwPassport... WWPassport) error {
	return nil
}

func (this wwPassportStorage) GetLastId(source string) (interface{}, error) {
	lastDate := time.Unix(0, 0)
	_, err := this.doFind(this.getLastIdQuery, func(rows *sql.Rows) error {
		rows.Scan(&lastDate)
		return nil
	}, source)
	if err != nil {
		return nil, err
	}
	return lastDate, nil
}