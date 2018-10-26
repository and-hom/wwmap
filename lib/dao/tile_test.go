package dao_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/DATA-DOG/go-sqlmock"
	"time"
	"database/sql"
	"database/sql/driver"
	"github.com/and-hom/wwmap/lib/model"
)

const POINT_JSON = "{\"type\":\"Point\",\"coordinates\":[100.121941666667,52.9252833333333]}"

func TestTileDaoZeroRows(t *testing.T) {
	tileDao, db := initDao(t)
	defer db.Close()

	found, err := tileDao.ListRiversWithBounds(geo.Bbox{}, false, 1)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(found))
}

func TestTileDaoSingleRiverSingleSpotNoImg(t *testing.T) {
	tileDao, db := initDao(t, []driver.Value{1, "Хара-Мурин", 1, "8", POINT_JSON, "4a", "http://aaa/bbb", "{}", -1, "", "", "", "", time.Unix(0, 0), ""})
	defer db.Close()

	found, err := tileDao.ListRiversWithBounds(geo.Bbox{}, false, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(found))
	assert.Equal(t, 1, len(found[0].Spots))
	assert.Equal(t, 0, len(found[0].Spots[0].Images))
}

func TestTileDaoSingleRiverSingleSpot(t *testing.T) {
	tileDao, db := initDao(t, []driver.Value{1, "Хара-Мурин", 1, "8", POINT_JSON, "4a", "http://aaa/bbb", "{}", 1, "wwmap", "1", "", "", time.Unix(0, 0), "image"})
	defer db.Close()

	found, err := tileDao.ListRiversWithBounds(geo.Bbox{}, false, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(found))
	assert.Equal(t, 1, len(found[0].Spots))
	assert.Equal(t, 1, len(found[0].Spots[0].Images))
}

func TestTileDaoTwoRiversTwoSpots(t *testing.T) {
	tileDao, db := initDao(t,
		[]driver.Value{1, "Хара-Мурин", 1, "8", POINT_JSON, "4a", "http://aaa/1", "{}", 1, "wwmap", "1", "", "", time.Unix(0, 0), "image"},
		[]driver.Value{1, "Хара-Мурин", 1, "8", POINT_JSON, "4a", "http://aaa/1", "{}", 2, "wwmap", "1", "", "", time.Unix(0, 0), "image"},
		[]driver.Value{1, "Хара-Мурин", 2, "12", POINT_JSON, "4a", "http://aaa/2", "{}", -1, "wwmap", "1", "", "", time.Unix(0, 0), "image"},
		[]driver.Value{2, "Жомболок", 3, "Катапульта", POINT_JSON, "5", "http://aaa/3", "{}", 4, "wwmap", "1", "", "", time.Unix(0, 0), "image"},
	)
	defer db.Close()

	found, err := tileDao.ListRiversWithBounds(geo.Bbox{}, false, 1)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(found))

	river0 := found[0]
	assert.Equal(t, int64(1), river0.Id)
	assert.Equal(t, "Хара-Мурин", river0.Title)
	assert.Equal(t, 2, len(river0.Spots))

	spot00 := river0.Spots[0]
	assert.Equal(t, int64(1), spot00.Id)
	assert.Equal(t, "8", spot00.Title)
	assert.Equal(t, 52.9252833333333, spot00.Point.Lat)
	assert.Equal(t, 100.121941666667, spot00.Point.Lon)
	assert.Equal(t, model.SportCategory{Category:4, Sub:"a"}, spot00.Category)
	assert.Equal(t, 2, len(spot00.Images))

	spot01 := river0.Spots[1]
	assert.Equal(t, int64(2), spot01.Id)
	assert.Equal(t, "12", spot01.Title)
	assert.Equal(t, 52.9252833333333, spot01.Point.Lat)
	assert.Equal(t, 100.121941666667, spot01.Point.Lon)
	assert.Equal(t, model.SportCategory{Category:4, Sub:"a"}, spot01.Category)
	assert.Equal(t, 0, len(spot01.Images))

	river1 := found[1]
	assert.Equal(t, int64(2), river1.Id)
	assert.Equal(t, "Жомболок", river1.Title)

	assert.Equal(t, 1, len(river1.Spots))
	spot10 := river1.Spots[0]
	assert.Equal(t, int64(3), spot10.Id)
	assert.Equal(t, "Катапульта", spot10.Title)
	assert.Equal(t, 52.9252833333333, spot10.Point.Lat)
	assert.Equal(t, 100.121941666667, spot10.Point.Lon)
	assert.Equal(t, model.SportCategory{Category:5, Sub:""}, spot10.Category)
	assert.Equal(t, 1, len(spot10.Images))
}

func initDao(t *testing.T, rowsData ...[]driver.Value) (dao.TileDao, *sql.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	rows := rowsHeader()
	for _, rd := range rowsData {
		rows = rows.AddRow(rd...)
	}

	mock.ExpectQuery(".*").WillReturnRows(rows);

	return dao.NewTilePostgresDao(dao.NewPostgresStorageForDb(db)), db
}

func rowsHeader() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"river_id", "river_title",
		"spot_id", "spot_title", "point", "category", "link", "props",
		"img_id", "img_source", "img_remote_id", "img_url", "img_preview_url", "img_date_published", "img_type"})
}
