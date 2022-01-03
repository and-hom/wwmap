package dao_test

import (
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/test"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var daoTester *test.DaoTester
var countryDao dao.CountryDao
var regionDao dao.RegionDao
var riverDao dao.RiverDao
var tileDao dao.TileDao

func TestMain(m *testing.M) {
	daoTester = &test.DaoTester{}
	daoTester.Init()

	postgresStorage := dao.NewPostgresStorageForDb(daoTester.Db)
	countryDao = dao.NewCountryPostgresDao(postgresStorage)
	regionDao = dao.NewRegionPostgresDao(postgresStorage)
	riverDao = dao.NewRiverPostgresDao(postgresStorage)
	tileDao = dao.NewTilePostgresDao(postgresStorage)

	log.Info("Dao initialized")

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := daoTester.Close(); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}
