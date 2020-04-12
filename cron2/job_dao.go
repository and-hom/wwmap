package main

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao"
)

type JobDao struct {
	dao.PostgresStorage
	insertQuery string
	updateQuery string
	listQuery   string
	getQuery    string
	deleteQuery string
}

func NewJobPostgresStorage(postgres dao.PostgresStorage) JobDao {
	return JobDao{
		PostgresStorage: postgres,
		insertQuery:     "INSERT INTO cron.job(title, expr, enabled, command, args) VALUES ($1,$2,$3,$4,$5) RETURNING id",
		updateQuery:     "UPDATE cron.job SET title=$2, expr=$3, enabled=$4, command=$5, args=$6 WHERE id=$1 RETURNING enabled<>(SELECT enabled FROM cron.job WHERE id=$1)",
		listQuery:       "SELECT id, title, expr, enabled, command, args FROM cron.job ORDER BY id DESC",
		getQuery:        "SELECT id, title, expr, enabled, command, args FROM cron.job WHERE id=$1",
		deleteQuery:     "DELETE FROM cron.job WHERE id=$1",
	}
}

func (this JobDao) get(id int64) (Job, bool, error) {
	p, found, err := this.DoFindAndReturn(this.getQuery, scanJob, id)
	if err != nil {
		return Job{}, false, err
	}
	if !found {
		return Job{}, false, nil
	}
	return p.(Job), true, nil
}

func (this JobDao) list() ([]Job, error) {
	lst, err := this.DoFindList(this.listQuery, scanJob)
	if err != nil {
		return []Job{}, err
	}
	return lst.([]Job), nil
}

func (this JobDao) remove(id int64) error {
	return this.PerformUpdatesWithinTxOptionally(nil, this.deleteQuery, dao.IdMapper, id)
}

func (this JobDao) insert(job Job) (int64, error) {
	id, err := this.UpdateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(Job)
		return []interface{}{_e.Title, _e.Expr, _e.Enabled, _e.Command, _e.Args}, nil
	}, true, job)
	if err != nil {
		return 0, err
	}
	return id[0], err
}

func (this JobDao) update(job Job) (bool, error) {
	result, err := this.UpdateReturningColumns(this.updateQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(Job)
		return []interface{}{_e.Id, _e.Title, _e.Expr, _e.Enabled, _e.Command, _e.Args}, nil
	}, true, job)
	if err != nil {
		return false, err
	}
	return *(result[0][0].(*bool)), err
}

func scanJob(rows *sql.Rows) (Job, error) {
	result := Job{}
	err := rows.Scan(&result.Id, &result.Title, &result.Expr, &result.Enabled, &result.Command, &result.Args)
	return result, err
}
