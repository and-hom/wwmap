package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
	"database/sql"
	"fmt"
)

func NewRegionPostgresDao(postgresStorage PostgresStorage) RegionDao {
	return regionStorage{
		PostgresStorage: postgresStorage,
		listQuery:queries.SqlQuery("region", "list-real"),
		listAllWithCountryQuery:queries.SqlQuery("region", "list-all-with-country"),
		getByIdQuery:queries.SqlQuery("region", "get-by-id"),
	}
}

type regionStorage struct {
	PostgresStorage
	getByIdQuery string
	listQuery    string
	listAllWithCountryQuery    string
}

func (this regionStorage) List(countryId int64) ([]Region, error) {
	lst, err := this.doFindList(this.listQuery, scanFunc, countryId)
	if err != nil {
		return []Region{}, err
	}
	return lst.([]Region), nil
}

func (this regionStorage) Get(id int64) (Region, error) {
	result, found, err := this.doFindAndReturn(this.getByIdQuery, scanFuncI, id)
	if err != nil {
		return Region{}, err
	}
	if !found {
		return Region{}, fmt.Errorf("Region with id %d not found", id)
	}
	return result.(Region), nil
}

func (this regionStorage) ListAllWithCountry() ([]RegionWithCountry, error) {
	lst, err := this.doFindList(this.listAllWithCountryQuery, func(rows *sql.Rows) (RegionWithCountry, error) {
		result := RegionWithCountry{
			Country: Country{},
		}
		err := rows.Scan(&result.Id, &result.Country.Id, &result.Country.Title, &result.Title)
		return result, err
	})
	if err != nil {
		return []RegionWithCountry{}, err
	}
	return lst.([]RegionWithCountry), nil
}

func scanFuncI(rows *sql.Rows) (interface{}, error) {
	return scanFunc(rows)
}
func scanFunc(rows *sql.Rows) (Region, error) {
	result := Region{}
	err := rows.Scan(&result.Id, &result.CountryId, &result.Title)
	return result, err
}

