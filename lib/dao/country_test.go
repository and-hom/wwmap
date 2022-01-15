package dao_test

import (
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestCountryList(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")

	countries, err := countryDao.List()

	assert.Nil(t, err)
	assert.ElementsMatch(t, []dao.Country{
		{
			Id:    260,
			Title: "Россия",
			Code:  "RU",
		},
		{
			Id:    261,
			Title: "Абхазия",
			Code:  "AB",
		},
		{
			Id:    55028,
			Title: "Казахстан",
			Code:  "KZ",
		},
	}, countries)
}

func TestCountryRemove(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")

	err := countryDao.Remove(261)

	assert.Nil(t, err)

	daoTester.TestDatabase(t, "country", "test/expected/country_after_remove.xml")
}

func TestCountryRemoveImpossibleWithRegions(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	daoTester.ApplyDbunitData(t, "test/region.xml")

	err := countryDao.Remove(261)

	assert.NotNil(t, err)

	daoTester.TestDatabase(t, "country", "test/country.xml")
}

func TestCountryInsert(t *testing.T) {
	ClearDb(t)
	id, err := countryDao.Insert(dao.Country{
		Code:  "RU",
		Title: "Россия",
	})

	assert.Nil(t, err)

	params := make(map[string]string)
	params["id"] = strconv.Itoa(int(id))
	daoTester.TestDatabase(t, "country", "test/expected/country_after_insert.xml", params)
}

func TestCountryInsertDuplicate(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	_, err := countryDao.Insert(dao.Country{
		Code:  "RU",
		Title: "Россия",
	})

	assert.NotNil(t, err)
	assert.IsType(t, dao.DuplicateError{}, err)
}


func TestCountryUpdate(t *testing.T) {
	ClearDb(t)
	daoTester.ApplyDbunitData(t, "test/country.xml")
	err := countryDao.Save(dao.Country{
		Id: 260,
		Code:  "BY",
		Title: "Беларусь",
	})

	assert.Nil(t, err)

	daoTester.TestDatabase(t, "country", "test/expected/country_after_update.xml")
}
