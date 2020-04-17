package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/util"
	"time"
)

func NewLevelPostgresDao(postgresStorage PostgresStorage) LevelDao {
	return &levelStorage{
		PostgresStorage:           postgresStorage,
		insertQuery:               queries.SqlQuery("level", "insert"),
		latestNotNullForDateQuery: queries.SqlQuery("level", "latest-not-null-for-date"),
		listQuery:                 queries.SqlQuery("level", "list-one"),
		listBySensorQuery:         queries.SqlQuery("level", "list-by-sensor"),
		removeNullsQuery:          queries.SqlQuery("level", "remove-nulls"),
	}
}

type levelStorage struct {
	PostgresStorage
	insertQuery               string
	latestNotNullForDateQuery string
	listQuery                 string
	listBySensorQuery         string
	removeNullsQuery          string
}

func (this levelStorage) Insert(entry Level) error {
	_, err := this.UpdateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(Level)
		levelValue := sql.NullInt64{
			Valid: _e.Level != NAN_LEVEL,
			Int64: int64(_e.Level),
		}
		return []interface{}{_e.SensorId, time.Time(_e.Date), _e.HourOfDay, levelValue}, nil
	}, true, entry)
	return err
}

func (this levelStorage) ListBySensorAndDate(fromDate time.Time, toDate time.Time) (map[string]map[string]Level, error) {
	lst, err := this.DoFindList(this.listQuery, scanLevel, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string]Level)
	for _, level := range lst.([]Level) {
		lvls, found := result[level.SensorId]
		if !found {
			lvls = make(map[string]Level)
		}
		t := util.ToDateInDefaultZone(time.Time(level.Date))
		lvls[util.FormatDate(t)] = level
		result[level.SensorId] = lvls
	}
	return result, nil
}

func (this levelStorage) RemoveNullsBefore(beforeDate JSONDate) error {
	return this.PerformUpdates(this.removeNullsQuery, dateToUpdateParams, time.Time(beforeDate))
}

func (this levelStorage) GetDailyLevelBetweenDates(sensorId string, from time.Time, to time.Time) ([]Level, error) {
	result, err := this.DoFindList(this.latestNotNullForDateQuery, scanLevel, sensorId, from, to)
	if err != nil {
		return []Level{}, err
	}
	return result.([]Level), err
}

func (this levelStorage) ListForSensor(sensorId string) ([]Level, error) {
	lst, err := this.DoFindList(this.listBySensorQuery, scanLevel, sensorId)
	if err != nil {
		return []Level{}, err
	}
	return lst.([]Level), nil
}

func dateToUpdateParams(date interface{}) ([]interface{}, error) {
	return []interface{}{date}, nil
}

func scanLevel(rows *sql.Rows) (Level, error) {
	result := Level{}
	err := rows.Scan(&result.Id, &result.SensorId, &result.Date, &result.HourOfDay, &result.Level)
	return result, err
}
