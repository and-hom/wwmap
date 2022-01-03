package dao_test

import (
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestRegionGetMissing(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	_, found, err := regionDao.Get(10001)
	assert.Nil(t, err)
	assert.False(t, found)
}

func TestRegionGet(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	region, found, err := regionDao.Get(268)

	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, dao.Region{
		Id:        268,
		CountryId: 260,
		Title:     "Алтай",
		Fake:      false,
	}, region)
}

func TestRegionGetFakeNotFound(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	_, found, err := regionDao.GetFake(260)

	assert.Nil(t, err)
	assert.False(t, found)
}

func TestRegionGetFake(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	region, found, err := regionDao.GetFake(261)

	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, dao.Region{
		Id:        269,
		CountryId: 261,
		Title:     "Абхазия",
		Fake:      true,
	}, region)
}

func TestRegionCreateFake(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	id, err := regionDao.CreateFake(260)
	assert.Nil(t, err)

	// Can't compare tables using dbuint because of random region name
	fake, found, err := regionDao.Get(id)
	assert.Nil(t, err)
	assert.True(t, found)
	assert.Equal(t, dao.Region{
		Id:        id,
		CountryId: 260,
		Title:     fake.Title,
		Fake:      true,
	}, fake)
}

func TestRegionCreateFakeDuplicate(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	_, err := regionDao.CreateFake(261)

	assert.NotNil(t, err)
}

func TestRegionList(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	regions, err := regionDao.List(260)

	assert.Nil(t, err)
	assert.ElementsMatch(t, []dao.Region{
		{
			Id:        268,
			CountryId: 260,
			Title:     "Алтай",
			Fake:      false,
		},
		{
			Id:        263,
			CountryId: 260,
			Title:     "Карелия",
			Fake:      false,
		},
	}, regions)
}
func TestRegionListAllWithCountry(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	regions, err := regionDao.ListAllWithCountry()

	assert.Nil(t, err)
	assert.ElementsMatch(t, []dao.RegionWithCountry{
		{
			Id:    268,
			Title: "Алтай",
			Fake:  false,
			Country: dao.Country{
				Id:    260,
				Title: "Россия",
				Code:  "RU",
			},
		},
		{
			Id:    263,
			Title: "Карелия",
			Fake:  false,
			Country: dao.Country{
				Id:    260,
				Title: "Россия",
				Code:  "RU",
			},
		},
		{
			Id:    269,
			Title: "Абхазия",
			Fake:  true,
			Country: dao.Country{
				Id:    261,
				Title: "Абхазия",
				Code:  "AB",
			},
		},
	}, regions)

}

func TestRegionSaveWrongCountry(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	err := regionDao.Save(dao.Region{
		Id:        269,
		CountryId: 0,
		Title:     "Новое название",
		Fake:      false,
	})
	assert.NotNil(t, err)

	daoTester.TestDatabase(t, "region", "test/region.xml")
}

func TestRegionSave(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	err := regionDao.Save(dao.Region{
		Id:        269,
		CountryId: 55028,
		Title:     "Новое название",
		Fake:      false,
	})
	assert.Nil(t, err)

	daoTester.TestDatabase(t, "region", "test/expected/region_after_save.xml")
}

func TestRegionInsertNonExistingCountry(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")

	_, err := regionDao.Insert(dao.Region{
		CountryId: 10001,
		Title:     "Мой регион",
		Fake:      false,
	})
	assert.NotNil(t, err)
}

func TestRegionInsert(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")

	id, err := regionDao.Insert(dao.Region{
		CountryId: 260,
		Title:     "Мой регион",
		Fake:      false,
	})
	assert.Nil(t, err)
	props := make(map[string]string)
	props["id"] = strconv.Itoa(int(id))
	daoTester.TestDatabase(t, "region", "test/expected/region_inserted.xml", props)
}

func TestRegionRemoveMissing(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	err := regionDao.Remove(10001)
	assert.Nil(t, err)

	daoTester.TestDatabase(t, "region", "test/region.xml")
}

func TestRegionRemove(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	err := regionDao.Remove(263)
	assert.Nil(t, err)

	daoTester.TestDatabase(t, "region", "test/expected/region_after_remove.xml")
}

func TestRegionRemoveInCountry(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	err := regionDao.RemoveAllByCountry(260)
	assert.Nil(t, err)

	daoTester.TestDatabase(t, "region", "test/expected/region_after_remove_all_in_country.xml")
}

func TestRegionGetParentIds(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	ids, err := regionDao.GetParentIds([]int64{268, 263, 269})
	assert.Nil(t, err)

	expected := make(map[int64]dao.RegionParentIds)
	expected[268] = dao.RegionParentIds{
		CountryId:   260,
		RegionTitle: "Алтай",
	}
	expected[263] = dao.RegionParentIds{
		CountryId:   260,
		RegionTitle: "Карелия",
	}
	expected[269] = dao.RegionParentIds{
		CountryId:   261,
		RegionTitle: "Абхазия",
	}
	assert.Equal(t, expected, ids)
}
