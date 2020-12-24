package dao

import (
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"time"
)

func NewChangesLogPostgresDao(postgresStorage PostgresStorage) ChangesLogDao {
	return &changesLogStorage{
		PostgresStorage:          postgresStorage,
		insertQuery:              queries.SqlQuery("changes-log", "insert"),
		listQuery:                queries.SqlQuery("changes-log", "list"),
		listAllQuery:             queries.SqlQuery("changes-log", "list-all"),
		listWithUserInfoQuery:    queries.SqlQuery("changes-log", "list-with-user-info"),
		listAllWithUserInfoQuery: queries.SqlQuery("changes-log", "list-all-with-user-info"),
		listAllTimeRangeQuery:    queries.SqlQuery("changes-log", "list-time-range"),
	}
}

type changesLogStorage struct {
	PostgresStorage
	insertQuery              string
	listQuery                string
	listAllQuery             string
	listWithUserInfoQuery    string
	listAllWithUserInfoQuery string
	listAllTimeRangeQuery    string
}

func (this changesLogStorage) Insert(entry ChangesLogEntry) error {
	_, err := this.UpdateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
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

func (this changesLogStorage) ListAllWithUserInfo(limit int) ([]ChangesLogEntryWithUseInfo, error) {
	lst, err := this.DoFindList(this.listAllWithUserInfoQuery, changesLogWithUserInfoMapper, limit)
	if err != nil {
		return []ChangesLogEntryWithUseInfo{}, err
	}
	return lst.([]ChangesLogEntryWithUseInfo), nil
}

func (this changesLogStorage) ListWithUserInfo(objectType string, objectId int64, limit int) ([]ChangesLogEntryWithUseInfo, error) {
	lst, err := this.DoFindList(this.listWithUserInfoQuery, changesLogWithUserInfoMapper, objectType, objectId, limit)
	if err != nil {
		return []ChangesLogEntryWithUseInfo{}, err
	}
	return lst.([]ChangesLogEntryWithUseInfo), nil
}

func (this changesLogStorage) ListAllTimeRange(fromInclude time.Time, toExclude time.Time, limit int) ([]ChangesLogEntry, error) {
	return this.list(this.listAllTimeRangeQuery, fromInclude, toExclude, limit)
}

func (this changesLogStorage) list(query string, args ...interface{}) ([]ChangesLogEntry, error) {
	lst, err := this.DoFindList(query, func(rows *sql.Rows) (ChangesLogEntry, error) {
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

func changesLogWithUserInfoMapper(rows *sql.Rows) (ChangesLogEntryWithUseInfo, error) {
	result := ChangesLogEntryWithUseInfo{}
	userInfo := ""
	err := rows.Scan(&result.Id, &result.ObjectType, &result.ObjectId, &result.AuthProvider, &result.ExtId,
		&result.ChangesLogEntry.Login, &result.Type, &result.Description, &result.Time, &userInfo)
	if err != nil {
		return ChangesLogEntryWithUseInfo{}, err
	}
	err = json.Unmarshal([]byte(userInfo), &result.UserInfo)
	return result, err
}