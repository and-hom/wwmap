package dao_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/and-hom/godbt"
	"github.com/and-hom/godbt/contract"
	cronDao "github.com/and-hom/wwmap/cron2/dao"
	daoLib "github.com/and-hom/wwmap/lib/dao"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/ory/dockertest"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"text/template"
)

var db *sql.DB
var tester *godbt.Tester
var dao cronDao.JobDao

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	log.Info("Connected to docker")

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgis/postgis", "latest", []string{
		"POSTGRES_DB=wwmap",
		"POSTGRES_USER=wwmap",
		"POSTGRES_PASSWORD=wwmap",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	port, err := strconv.Atoi(resource.GetPort("5432/tcp"))
	if err != nil {
		log.Fatalf("Could not parse port %s: %s", resource.GetPort("5432/tcp"), err)
	}
	log.Infof("Postgres started on port %d", port)

	connString := fmt.Sprintf("postgres://wwmap:wwmap@localhost:%d/wwmap?sslmode=disable", port)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", connString)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	log.Info("Connected to database")

	absMigrationsPath, err := filepath.Abs("../../db")
	if err != nil {
		log.Fatalf("Could get migrations path: %s", err)
	}
	log.Infof("Loading migrations from %s", absMigrationsPath)
	migrations, err := migrate.New(
		"file://"+absMigrationsPath,
		connString)
	if err != nil {
		log.Fatalf("Could load migrations: %s", err)
	}
	err = migrations.Up()
	if err != nil {
		log.Fatalf("Could apply migrations: %s", err)
	}
	log.Info("Migrations applied")

	tester, err = godbt.GetTester(contract.InstallerConfig{
		Type:        "postgres",
		ConnString:  connString,
		ClearMethod: contract.ClearMethodDeleteAll,
	})
	if err != nil {
		log.Fatalf("Could init dbunit: %s", err)
	}
	log.Info("DBUnit initialized")

	dao = cronDao.NewJobPostgresStorage(daoLib.NewPostgresStorageForDb(db))

	log.Info("Dao initialized")

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestSelectJobsEmpty(t *testing.T) {
	clearTable(t, "cron.job")
	result, err := dao.List()
	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestSelectJobs(t *testing.T) {
	applyDbunitData(t, "test/jobs.xml")

	result, err := dao.List()
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
	applyDbunitData(t, "test/jobs.xml")

	_, found, err := dao.Get(1002)
	assert.Nil(t, err)
	assert.False(t, found)
}

func TestSelectSingleJob(t *testing.T) {
	applyDbunitData(t, "test/jobs.xml")

	result, found, err := dao.Get(1001)
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
	applyDbunitData(t, "test/jobs.xml")

	err := dao.Remove(1002)
	assert.Nil(t, err)
	testDatabase(t, "cron.job", "test/jobs.xml")
}

func TestRemoveJob(t *testing.T) {
	applyDbunitData(t, "test/jobs.xml")

	err := dao.Remove(1001)
	assert.Nil(t, err)
	testDatabase(t, "cron.job", "test/expected/jobs_empty.xml")
}

func TestInsertJob(t *testing.T) {
	clearTable(t, "cron.job")

	id, err := dao.Insert(cronDao.Job{
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
	testDatabase(t, "cron.job", "test/expected/jobs_inserted.xml", params)
}

func TestUpdateMissing(t *testing.T) {
	applyDbunitData(t, "test/jobs.xml")

	_, err := dao.Update(cronDao.Job{
		daoLib.IdTitle{1002, "Job2"},
		"*/2 * * * *",
		false,
		true,
		"another_command",
		"args",
	})
	assert.NotNil(t, err)
	testDatabase(t, "cron.job", "test/jobs.xml")
}

func TestUpdateJobEnabledChanged(t *testing.T) {
	applyDbunitData(t, "test/jobs.xml")

	scheduleChanged, err := dao.Update(cronDao.Job{
		daoLib.IdTitle{1001, "New name"},
		"*/1 * * * *",
		false,
		true,
		"another_command",
		"args",
	})
	assert.Nil(t, err)
	assert.True(t, scheduleChanged)
	testDatabase(t, "cron.job", "test/expected/jobs_enabled_changed.xml")
}

func TestUpdateJobCronChanged(t *testing.T) {
	applyDbunitData(t, "test/jobs.xml")

	scheduleChanged, err := dao.Update(cronDao.Job{
		daoLib.IdTitle{1001, "New name"},
		"*/3 * * * *",
		true,
		true,
		"another_command",
		"args",
	})
	assert.Nil(t, err)
	assert.True(t, scheduleChanged)
	testDatabase(t, "cron.job", "test/expected/jobs_cron_changed.xml")
}

func TestUpdateJob(t *testing.T) {
	applyDbunitData(t, "test/jobs.xml")

	scheduleChanged, err := dao.Update(cronDao.Job{
		daoLib.IdTitle{1001, "New name"},
		"*/1 * * * *",
		true,
		true,
		"another_command",
		"args",
	})
	assert.Nil(t, err)
	assert.False(t, scheduleChanged)
	testDatabase(t, "cron.job", "test/expected/jobs_changed_other_fields.xml")
}

func clearTable(t *testing.T, table string) {
	_, err := db.Exec("DELETE FROM " + table)
	if err != nil {
		assert.Fail(t, "Can't clear table %s: %s", table, err)
	}
}

func applyDbunitData(t *testing.T, path string) {
	image := loadImg(t, path)
	err := tester.GetInstaller().InstallImage(image)
	if err != nil {
		assert.Fail(t, "Can't apply dbunit data from %s: %s", path, err)
	}
}

func testDatabase(t *testing.T, table string, path string, opts ...interface{}) {
	var params map[string]string
	if len(opts) > 0 {
		ok := false
		params, ok = opts[0].(map[string]string)
		if !ok {
			assert.Fail(t, "Param should be map[string]string but was %v", opts[0])
		}
	} else {
		params = make(map[string]string)
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		assert.Fail(t, "Can't load file %s: %s", path, err)
	}

	tmpl, err := template.New("replace").Parse(string(content))
	if err != nil {
		assert.Fail(t, "Can't load template: %s", err)
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, params)
	if err != nil {
		assert.Fail(t, "Can't apply template: %s", err)
	}


	expectedImage, err := tester.GetImageManager().LoadImage(buf.String())
	if err != nil {
		assert.Fail(t, "Can't load dbunit data from xml %s: %s", buf.String(), err)
	}
	actualImage, err := tester.GetInstaller().GetTableImage(table)
	if err != nil {
		assert.Fail(t, "Can't load table data from %s: %s", path, err)
	}
	diffs := tester.GetImageManager().GetImagesDiff(expectedImage, actualImage)
	if len(diffs) > 0 {
		assert.Fail(t, "Tables are different: %v", diffs)
	}
}

func loadImg(t *testing.T, path string) contract.Image {
	image, err := tester.GetImageManager().LoadImage(path)
	if err != nil {
		assert.Fail(t, "Can't load dbunit data from %s: %s", path, err)
	}
	return image
}
