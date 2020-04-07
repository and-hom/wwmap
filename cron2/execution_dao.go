package main

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/lib/pq"
	"time"
)

type ExecutionDao struct {
	dao.PostgresStorage
	listQuery string
}

func NewExecutionPostgresStorage(postgres dao.PostgresStorage) ExecutionDao {
	return ExecutionDao{
		PostgresStorage: postgres,
		listQuery:       "SELECT id, job_id, start, \"end\", status FROM cron.execution WHERE \"end\">=$1 AND start<$2 OR \"end\" IS NULL ORDER BY start ASC",
	}
}

func (this ExecutionDao) list(from time.Time, to time.Time) ([]Execution, error) {
	lst, err := this.DoFindList(this.listQuery, scanExecution, from, to)
	if err != nil {
		return []Execution{}, err
	}
	return lst.([]Execution), nil
}

func scanExecution(rows *sql.Rows) (Execution, error) {
	result := Execution{}
	end := pq.NullTime{}
	err := rows.Scan(&result.Id, &result.JobId, &result.Start, &end, &result.Status)
	if end.Valid {
		t := dao.JSONTime(end.Time)
		result.End = &t
	} else {
		result.End = nil
	}
	return result, err
}
