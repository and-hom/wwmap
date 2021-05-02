package dao_test

import (
	cronDao "github.com/and-hom/wwmap/cron2/dao"
	daoLib "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/test"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

var daoTester *DaoTester
var jobDao cronDao.JobDao
var executionDao cronDao.ExecutionDao

func TestMain(m *testing.M) {
	daoTester = &DaoTester{}
	daoTester.Init()
	jobDao = cronDao.NewJobPostgresStorage(daoLib.NewPostgresStorageForDb(daoTester.Db))
	executionDao = cronDao.NewExecutionPostgresStorage(daoLib.NewPostgresStorageForDb(daoTester.Db))

	log.Info("Dao initialized")

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := daoTester.Close(); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestSelectJobsEmpty(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ClearTable(t, "cron.job")
	result, err := jobDao.List()
	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestSelectJobs(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	result, err := jobDao.List()
	assert.Nil(t, err)
	assert.Equal(t, result, []cronDao.Job{
		cronDao.Job{
			daoLib.IdTitle{1001, "Job"},
			"*/1 * * * *",
			true,
			false,
			"command",
			"arg1 arg2",
		},
	})
}

func TestSelectSingleJobBadId(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	_, found, err := jobDao.Get(1002)
	assert.Nil(t, err)
	assert.False(t, found)
}

func TestSelectSingleJob(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	result, found, err := jobDao.Get(1001)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, result, cronDao.Job{
		daoLib.IdTitle{1001, "Job"},
		"*/1 * * * *",
		true,
		false,
		"command",
		"arg1 arg2",
	})
}

func TestRemoveJobBadId(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	err := jobDao.Remove(1002)
	assert.Nil(t, err)
	daoTester.TestDatabase(t, "cron.job", "test/jobs.xml")
}

func TestRemoveJob(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	err := jobDao.Remove(1001)
	assert.Nil(t, err)
	daoTester.TestDatabase(t, "cron.job", "test/expected/empty.xml")
}

func TestInsertJob(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ClearTable(t, "cron.job")

	id, err := jobDao.Insert(cronDao.Job{
		daoLib.IdTitle{1002, "Job2"},
		"*/2 * * * *",
		false,
		true,
		"another_command",
		"args",
	})
	assert.Nil(t, err)
	params := make(map[string]string)
	params["id"] = strconv.Itoa(int(id))
	daoTester.TestDatabase(t, "cron.job", "test/expected/jobs_inserted.xml", params)
}

func TestUpdateMissing(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	_, err := jobDao.Update(cronDao.Job{
		daoLib.IdTitle{1002, "Job2"},
		"*/2 * * * *",
		false,
		true,
		"another_command",
		"args",
	})
	assert.NotNil(t, err)
	daoTester.TestDatabase(t, "cron.job", "test/jobs.xml")
}

func TestUpdateJobEnabledChanged(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	scheduleChanged, err := jobDao.Update(cronDao.Job{
		daoLib.IdTitle{1001, "New name"},
		"*/1 * * * *",
		false,
		true,
		"another_command",
		"args",
	})
	assert.Nil(t, err)
	assert.True(t, scheduleChanged)
	daoTester.TestDatabase(t, "cron.job", "test/expected/jobs_enabled_changed.xml")
}

func TestUpdateJobCronChanged(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	scheduleChanged, err := jobDao.Update(cronDao.Job{
		daoLib.IdTitle{1001, "New name"},
		"*/3 * * * *",
		true,
		true,
		"another_command",
		"args",
	})
	assert.Nil(t, err)
	assert.True(t, scheduleChanged)
	daoTester.TestDatabase(t, "cron.job", "test/expected/jobs_cron_changed.xml")
}

func TestUpdateJob(t *testing.T) {
	daoTester.ClearTable(t, "cron.execution")
	daoTester.ApplyDbunitData(t, "test/jobs.xml")

	scheduleChanged, err := jobDao.Update(cronDao.Job{
		daoLib.IdTitle{1001, "New name"},
		"*/1 * * * *",
		true,
		true,
		"another_command",
		"args",
	})
	assert.Nil(t, err)
	assert.False(t, scheduleChanged)
	daoTester.TestDatabase(t, "cron.job", "test/expected/jobs_changed_other_fields.xml")
}
