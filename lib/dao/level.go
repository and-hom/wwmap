package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"time"
)

func NewLevelPostgresDao(postgresStorage PostgresStorage) LevelDao {
	return &levelStorage{
		PostgresStorage:  postgresStorage,
		insertQuery:      queries.SqlQuery("level", "insert"),
		listQuery:        queries.SqlQuery("level", "list-one"),
		removeNullsQuery: queries.SqlQuery("level", "remove-nulls"),
	}
}

type levelStorage struct {
	PostgresStorage
	insertQuery      string
	listQuery        string
	removeNullsQuery string
}

func (this levelStorage) Insert(entry Level) error {
	_, err := this.updateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(Level)
		levelValue := sql.NullInt64{
			Valid: _e.Level != NAN_LEVEL,
			Int64: int64(_e.Level),
		}
		return []interface{}{_e.SensorId, time.Time(_e.Date), _e.HourOfDay, levelValue}, nil
	}, true, entry)
	return err
}

func (this levelStorage) List(fromDate JSONDate) (map[string][]Level, error) {
	lst, err := this.doFindList(this.listQuery, scanLevel, time.Time(fromDate))
	if err != nil {
		return nil, err
	}
	result := make(map[string][]Level)
	for _, level := range lst.([]Level) {
		result[level.SensorId] = append(result[level.SensorId], level)
	}
	return result, nil
}

func (this levelStorage) RemoveNullsBefore(beforeDate JSONDate) error {
	return this.performUpdates(this.removeNullsQuery, dateToUpdateParams, time.Time(beforeDate))
}

func dateToUpdateParams(date interface{}) ([]interface{}, error) {
	return []interface{}{date}, nil
}

func scanLevel(rows *sql.Rows) (Level, error) {
	result := Level{}
	err := rows.Scan(&result.Id, &result.SensorId, &result.Date, &result.HourOfDay, &result.Level)
	return result, err
}
