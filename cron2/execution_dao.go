package main

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/lib/pq"
	"time"
)

type ExecutionDao struct {
	dao.PostgresStorage
	getQuery                 string
	listQuery                string
	insertQuery              string
	updateStatusQuery        string
	deleteByJobQuery         string
	markRunningAsOrphanQuery string
}

func NewExecutionPostgresStorage(postgres dao.PostgresStorage) ExecutionDao {
	return ExecutionDao{
		PostgresStorage:          postgres,
		getQuery:                 "SELECT id, job_id, start, \"end\", status FROM cron.execution WHERE id=$1",
		listQuery:                "SELECT id, job_id, start, \"end\", status FROM cron.execution WHERE \"end\">=$1 AND start<$2 OR \"end\" IS NULL ORDER BY start ASC",
		insertQuery:              "INSERT INTO cron.execution(job_id, start, \"end\", status) VALUES ($1, $2, $3, $4) RETURNING id, job_id, start, COALESCE(\"end\", to_timestamp(0)), status",
		updateStatusQuery:        "UPDATE cron.execution SET \"end\" = $2, status = $3 WHERE id=$1",
		deleteByJobQuery:         "DELETE FROM cron.execution WHERE job_id=$1",
		markRunningAsOrphanQuery: "UPDATE cron.execution SET status=$2, \"end\"=$3 WHERE status=$1",
	}
}

func (this ExecutionDao) get(id int64) (Execution, bool, error) {
	p, found, err := this.DoFindAndReturn(this.getQuery, scanExecution, id)
	if err != nil {
		return Execution{}, false, err
	}
	if !found {
		return Execution{}, false, nil
	}
	return p.(Execution), true, nil
}

func (this ExecutionDao) list(from time.Time, to time.Time) ([]Execution, error) {
	lst, err := this.DoFindList(this.listQuery, scanExecution, from, to)
	if err != nil {
		return []Execution{}, err
	}
	return lst.([]Execution), nil
}

func (this ExecutionDao) insert(jobId int64) (Execution, error) {
	cols, err := this.UpdateReturningColumns(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		jId := entity.(int64)
		return []interface{}{jId, time.Now(), pq.NullTime{}, RUNNING}, nil
	}, true, jobId)
	if err != nil {
		return Execution{}, err
	}

	tStart := cols[0][2].(*time.Time)
	tEnd := cols[0][3].(*time.Time)
	var tEndPtr *dao.JSONTime = nil
	if !tEnd.Before(*tStart) {
		tEndJson := dao.JSONTime(*tEnd)
		tEndPtr = &tEndJson

	}

	return Execution{
		Id:     *(cols[0][0].(*int64)),
		JobId:  *(cols[0][1].(*int64)),
		Start:  dao.JSONTime(*tStart),
		End:    tEndPtr,
		Status: Status(*(cols[0][4].(*string))),
	}, err
}

func (this ExecutionDao) setStatus(id int64, status Status) error {
	return this.PerformUpdates(this.updateStatusQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.([]interface{})
		return []interface{}{_e[0], time.Now(), _e[1]}, nil
	}, []interface{}{id, status})
}

func (this ExecutionDao) removeByJob(jobId int64) error {
	return this.PerformUpdates(this.deleteByJobQuery, dao.IdMapper, jobId)
}

func (this ExecutionDao) markRunningAsOrphan() error {
	return this.PerformUpdates(this.markRunningAsOrphanQuery, dao.ArrayMapper, []interface{}{RUNNING, ORPHAN, time.Now()})
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
