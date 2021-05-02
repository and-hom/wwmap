package dao_test

import (
	cronDao "github.com/and-hom/wwmap/cron2/dao"
	"github.com/and-hom/wwmap/lib/dao"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestSelectExecutionByIdNotFound(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	_, found, err := executionDao.Get(1005)

	assert.Nil(t, err)
	assert.False(t, found)
}

func TestSelectExecutionById(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	result, found, err := executionDao.Get(1001)
	date := time.Date(2020, 12, 10, 10, 0, 0, 0, time.UTC)
	expected := cronDao.Execution{
		Id:     1001,
		JobId:  1001,
		Start:  dao.JSONTime(date),
		End:    nil,
		Status: cronDao.RUNNING,
		Manual: false,
	}

	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, result, expected)
}

func TestListSorted(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	t_07_00 := time.Date(2020, 12, 10, 7, 0, 0, 0, time.UTC)
	t_08_00 := time.Date(2020, 12, 10, 8, 0, 0, 0, time.UTC)
	t_08_01 := time.Date(2020, 12, 10, 8, 1, 0, 0, time.UTC)
	t_10_00 := time.Date(2020, 12, 10, 10, 0, 0, 0, time.UTC)

	t_08_01_json := dao.JSONTime(t_08_01)

	executions, err := executionDao.ListSorted(t_07_00, t_08_01)

	assert.Nil(t, err)
	assert.Equal(t, []cronDao.Execution{
		{
			Id:     1004,
			JobId:  1001,
			Start:  dao.JSONTime(t_07_00),
			End:    nil,
			Status: cronDao.ORPHAN,
			Manual: false,
		},
		{
			Id:     1003,
			JobId:  1001,
			Start:  dao.JSONTime(t_08_00),
			End:    &t_08_01_json,
			Status: cronDao.FAIL,
			Manual: false,
		},
		{
			Id:     1001,
			JobId:  1001,
			Start:  dao.JSONTime(t_10_00),
			End:    nil,
			Status: cronDao.RUNNING,
			Manual: false,
		},
	}, executions)
}

func TestInsert(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	execution, err := executionDao.Insert(int64(1001), true)

	expected := cronDao.Execution{
		Id:     execution.Id, // ignore comparsion of id
		JobId:  1001,
		Start:  execution.Start, // ignore comparsion of date
		End:    nil,
		Status: cronDao.RUNNING,
		Manual: true,
	}

	assert.Nil(t, err)
	assert.Equal(t, execution, expected)

	params := make(map[string]string)
	params["id"] = strconv.Itoa(int(execution.Id))
	params["start"] = time.Time(execution.Start).Format("2006-01-02T15:04:05.000000Z")
	daoTester.TestDatabase(t, "cron.execution", "test/expected/execution.xml", params)
}

func TestSetStatusMissing(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	err := executionDao.SetStatus(1006, cronDao.FAIL, time.Now())
	assert.Nil(t, err)
	daoTester.TestDatabase(t, "cron.execution", "test/expected/executions.xml")
}

func TestSetStatus(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	date := time.Date(2020, 12, 10, 10, 1, 0, 0, time.UTC)
	err := executionDao.SetStatus(1001, cronDao.FAIL, date)
	assert.Nil(t, err)
	daoTester.TestDatabase(t, "cron.execution", "test/expected/executions_status_changed.xml")
}

func TestRemoveByJobMissing(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	err := executionDao.RemoveByJob(1002)
	assert.Nil(t, err)
	daoTester.TestDatabase(t, "cron.execution", "test/expected/executions.xml")
}

func TestRemoveByJob(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	err := executionDao.RemoveByJob(1001)
	assert.Nil(t, err)
	daoTester.TestDatabase(t, "cron.execution", "test/expected/empty.xml")
}

func TestRemoveOld1(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	date := time.Date(2020, 12, 10, 9, 1, 0, 0, time.UTC)
	maxId, count, err := executionDao.RemoveOld(date)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, int64(1003), maxId)
	daoTester.TestDatabase(t, "cron.execution", "test/expected/executions_after_clean_1.xml")
}

func TestRemoveOld2(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	date := time.Date(2020, 12, 10, 9, 2, 0, 0, time.UTC)
	maxId, count, err := executionDao.RemoveOld(date)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), count)
	assert.Equal(t, int64(1003), maxId)
	daoTester.TestDatabase(t, "cron.execution", "test/expected/executions_after_clean_2.xml")
}

func TestMarkRunningAsOrphan(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")
	daoTester.ApplyDbunitData(t, "test/executions.xml")

	date := time.Date(2021, 5, 1, 0, 0, 0, 0, time.UTC)

	err := executionDao.MarkRunningAsOrphan(date)
	assert.Nil(t, err)

	daoTester.TestDatabase(t, "cron.execution", "test/expected/executions_running_as_orphan.xml")
}