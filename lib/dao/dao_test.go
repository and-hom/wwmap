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
var imgDao dao.ImgDao

func TestMain(m *testing.M) {
	daoTester = &test.DaoTester{}
	daoTester.Init()

	postgresStorage := dao.NewPostgresStorageForDb(daoTester.Db)
	countryDao = dao.NewCountryPostgresDao(postgresStorage)
	regionDao = dao.NewRegionPostgresDao(postgresStorage)
	riverDao = dao.NewRiverPostgresDao(postgresStorage)
	tileDao = dao.NewTilePostgresDao(postgresStorage)
	imgDao = dao.NewImgPostgresDao(postgresStorage)

	log.Info("Dao initialized")

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := daoTester.Close(); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func ClearDb(t *testing.T) {
	daoTester.ClearTable(t, "image")
	daoTester.ClearTable(t, "voyage_report_river")
	daoTester.ClearTable(t, "voyage_report")
	daoTester.ClearTable(t, "white_water_rapid")
	daoTester.ClearTable(t, "river")
	daoTester.ClearTable(t, "region")
	daoTester.ClearTable(t, "country")
}
