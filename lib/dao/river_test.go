package dao_test

import (
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

var expectedRiver1 = dao.River{
	RiverTitle: dao.RiverTitle{
		IdTitle: dao.IdTitle{
			Id:    1,
			Title: "Умба",
		},
		OsmId: 0,
		Region: dao.Region{
			Id:        263,
			CountryId: 260,
			Title:     "Карелия",
			Fake:      false,
		},
		Bounds:  geo.Bbox{},
		Aliases: []string{},
		Props:   make(map[string]interface{}),
		Visible: true,
	},
	Description:  "Река",
	SpotCounters: dao.SpotCounters{
		Ordered: 0,
		Total:   1,
	},
}

func TestRiverInsert(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	props := make(map[string]interface{})
	props["a"] = 1
	props["b"] = "b"
	aliases := []string{"Riv 1", "R1"}

	id, err := riverDao.Insert(dao.River{
		RiverTitle: dao.RiverTitle{
			IdTitle: dao.IdTitle{Title: "River 1"},
			Region:  dao.Region{Id: 268},
			Aliases: aliases,
			Props:   props,
		},
		Description: "Description",
	})

	assert.Nil(t, err)
	p := make(map[string]string)
	p["id"] = strconv.Itoa(int(id))
	daoTester.TestDatabase(t, "river", "test/expected/river_inserted.xml", p, []string{"last_modified"})
}

func TestRiverSave(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")

	props1 := make(map[string]interface{})
	props1["a"] = "b"
	props2 := make(map[string]interface{})
	props2["a"] = 1

	err := riverDao.Save(
		dao.RiverTitle{
			IdTitle: dao.IdTitle{Id: 1, Title: "Умба"},
			Region:  dao.Region{Id: 263},
			Aliases: []string{},
			Props:   props1,
			Visible: false,
		},
		dao.RiverTitle{
			IdTitle: dao.IdTitle{Id: 2, Title: "Кодор"},
			Region:  dao.Region{Id: 269},
			Props:   props2,
			Visible: true,
		},
	)

	assert.Nil(t, err)

	daoTester.TestDatabase(
		t,
		"river",
		"test/expected/river_updated.xml",
		make(map[string]string),
		[]string{"last_modified"},
	)
}

func TestRiverSaveFull(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")

	props1 := make(map[string]interface{})
	props1["a"] = "b"
	props2 := make(map[string]interface{})
	props2["a"] = 1

	err := riverDao.SaveFull(
		dao.River{
			RiverTitle: dao.RiverTitle{
				IdTitle: dao.IdTitle{Id: 1, Title: "Умба"},
				Region:  dao.Region{Id: 263},
				Aliases: []string{},
				Props:   props1,
				Visible: false,
			},
			Description: "Река Умба",
		},
		dao.River{
			RiverTitle: dao.RiverTitle{
				IdTitle: dao.IdTitle{Id: 2, Title: "Кодор"},
				Region:  dao.Region{Id: 269},
				Props:   props2,
				Visible: true,
			},
			Description: "Река Кодор",
		},
	)

	assert.Nil(t, err)

	daoTester.TestDatabase(
		t,
		"river",
		"test/expected/river_updated_full.xml",
		make(map[string]string),
		[]string{"last_modified"},
	)
}

func TestRiverSetVisible(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")

	err1 := riverDao.SetVisible(1, false)
	assert.Nil(t, err1)

	err2 := riverDao.SetVisible(2, true)
	assert.Nil(t, err2)

	daoTester.TestDatabase(
		t,
		"river",
		"test/expected/river_visibility_change.xml",
		make(map[string]string),
		[]string{"last_modified"},
	)
}

func TestRiverFindMissing(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")

	_, err := riverDao.Find(100)
	assert.NotNil(t, err)
	nfe, ok := err.(dao.EntityNotFoundError)
	assert.True(t, ok)
	assert.Equal(t, dao.RIVER, nfe.EntityType)
	assert.Equal(t, int64(100), nfe.Id)
}

func TestRiverFind(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")

	river, err := riverDao.Find(1)
	assert.Nil(t, err)

	assert.Equal(t, expectedRiver1, river)
}

func TestRiverFindForImage(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")
	daoTester.ApplyDbunitData(t, "test/voyage_report.xml")
	daoTester.ApplyDbunitData(t, "test/image.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid_image.xml")

	river, err := riverDao.FindForImage(1)
	assert.Nil(t, err)

	assert.Equal(t, expectedRiver1, river)
}

func TestRiverFindForSpot(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")
	daoTester.ApplyDbunitData(t, "test/white_water_rapid.xml")

	river, err := riverDao.FindForSpot(1)
	assert.Nil(t, err)

	assert.Equal(t, expectedRiver1, river)
}

func TestRiverCountByCountry(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")

	count, err := riverDao.CountByCountry(260)
	assert.Nil(t, err)

	assert.Equal(t, 2, count)
}

func TestRiverCountByCountry0(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")
	daoTester.ApplyDbunitData(t, "test/river.xml")

	count, err := riverDao.CountByCountry(261)
	assert.Nil(t, err)

	assert.Equal(t, 0, count)
}
