package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"time"
)

func NewLevelPostgresDao(postgresStorage PostgresStorage) LevelDao {
	return &levelStorage{
		PostgresStorage: postgresStorage,
		insertQuery:     queries.SqlQuery("level", "insert"),
	}
}

type levelStorage struct {
	PostgresStorage
	insertQuery string
}

func (this levelStorage) Insert(entry Level) error {
	_, err := this.updateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(Level)
		levelValue := sql.NullInt64{
			Valid: _e.Level != NAN_LEVEL,
			Int64: int64(_e.Level),
		}
		return []interface{}{_e.SensorId, time.Time(_e.Date), _e.HourOfDay, levelValue}, nil
	}, entry)
	return err
}
