package test

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/and-hom/godbt"
	"github.com/and-hom/godbt/contract"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/ory/dockertest"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"
	"text/template"
)

type DaoTester struct {
	Db       *sql.DB
	Tester   *godbt.Tester
	Resource *dockertest.Resource
	Pool     *dockertest.Pool
}

func (this *DaoTester) Init() {
	var err error
	this.Pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	log.Info("Connected to docker")

	// pulls an image, creates a container based on it and runs it
	this.Resource, err = this.Pool.Run("postgis/postgis", "latest", []string{
		"POSTGRES_DB=wwmap",
		"POSTGRES_USER=wwmap",
		"POSTGRES_PASSWORD=wwmap",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	port, err := strconv.Atoi(this.Resource.GetPort("5432/tcp"))
	if err != nil {
		log.Fatalf("Could not parse port %s: %s", this.Resource.GetPort("5432/tcp"), err)
	}
	log.Infof("Postgres started on port %d", port)

	connString := fmt.Sprintf("postgres://wwmap:wwmap@localhost:%d/wwmap?sslmode=disable", port)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := this.Pool.Retry(func() error {
		var err error
		this.Db, err = sql.Open("postgres", connString)
		if err != nil {
			return err
		}
		return this.Db.Ping()
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

	this.Tester, err = godbt.GetTester(contract.InstallerConfig{
		Type:        "postgres",
		ConnString:  connString,
		ClearMethod: contract.ClearMethodDeleteAll,
	})
	if err != nil {
		log.Fatalf("Could init dbunit: %s", err)
	}
	log.Info("DBUnit initialized")
}

func (this *DaoTester) Close() error {
	return this.Pool.Purge(this.Resource)
}

func (this *DaoTester) ClearTable(t *testing.T, table string) {
	_, err := this.Db.Exec("DELETE FROM " + table)
	if err != nil {
		assert.Fail(t, "Can't clear table %s: %s", table, err)
	}
}

func (this *DaoTester) ApplyDbunitData(t *testing.T, path string) {
	image := this.loadImg(t, path)
	err := this.Tester.GetInstaller().InstallImage(image)
	if err != nil {
		assert.Fail(t, "Can't apply dbunit data from %s: %s", path, err)
	}
}

func (this *DaoTester) TestDatabase(t *testing.T, table string, path string, opts ...interface{}) {
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

	expectedImage, err := this.Tester.GetImageManager().LoadImage(buf.String())
	if err != nil {
		assert.Fail(t, "Can't load dbunit data from xml %s: %s", buf.String(), err)
	}
	actualImage, err := this.Tester.GetInstaller().GetTableImage(table)
	if err != nil {
		assert.Fail(t, "Can't load table data from %s: %s", path, err)
	}
	diffs := this.Tester.GetImageManager().GetImagesDiff(expectedImage, actualImage)
	if len(diffs) > 0 {
		assert.Fail(t, "Tables are different: %v", diffs)
	}
}

func (this *DaoTester) loadImg(t *testing.T, path string) contract.Image {
	image, err := this.Tester.GetImageManager().LoadImage(path)
	if err != nil {
		assert.Fail(t, "Can't load dbunit data from %s: %s", path, err)
	}
	return image
}
