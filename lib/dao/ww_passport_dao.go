package dao

import (
	"time"
	"database/sql"
)

type WWPassportStorage struct {
	PostgresStorage
}

func (this WWPassportStorage) Upsert(wwPassport... WWPassport) error {
	return nil
}

func (this WWPassportStorage) GetLastId(source string) (interface{}, error) {
	lastDate := time.Unix(0, 0)
	_, err := this.doFind("SELECT max(date_modified) FROM ww_passport WHERE source=$1", func(rows *sql.Rows) error {
		rows.Scan(&lastDate)
		return nil
	}, source)
	if err != nil {
		return nil, err
	}
	return lastDate, nil
}