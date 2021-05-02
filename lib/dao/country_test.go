package dao_test

import (
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCountryList(t *testing.T) {
	daoTester.ClearTable(t, "region")
	daoTester.ClearTable(t, "country")
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
