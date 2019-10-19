package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
	"time"
)

func NewMeteoPostgresDao(postgresStorage PostgresStorage) MeteoDao {
	return &meteoStorage{
		PostgresStorage: postgresStorage,
		insertQuery:     queries.SqlQuery("meteo", "insert"),
	}
}

type meteoStorage struct {
	PostgresStorage
	insertQuery string
}

func (this meteoStorage) Insert(entry Meteo) error {
	_, err := this.updateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(Meteo)
		return []interface{}{_e.PointId, time.Time(_e.Date), string(_e.Daytime), _e.Temp, _e.Rain}, nil
	}, true, entry)
	return err
}
