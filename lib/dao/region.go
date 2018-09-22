package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
	"database/sql"
	"fmt"
)

func NewRegionPostgresDao(postgresStorage PostgresStorage) RegionDao {
	return regionStorage{
		PostgresStorage: postgresStorage,
		PropsManager:PropertyManagerImpl{table:queries.SqlQuery("region", "table"), dao:&postgresStorage},
		listQuery:queries.SqlQuery("region", "list-real"),
		listAllWithCountryQuery:queries.SqlQuery("region", "list-all-with-country"),
		getByIdQuery:queries.SqlQuery("region", "get-by-id"),
		getFakeQuery:queries.SqlQuery("region", "get-fake"),
		createFakeQuery:queries.SqlQuery("region", "create-fake"),
	}
}

type regionStorage struct {
	PostgresStorage
	PropsManager            PropertyManager
	getByIdQuery            string
	listQuery               string
	listAllWithCountryQuery string
	getFakeQuery            string
	createFakeQuery         string
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
		err := rows.Scan(&result.Id, &result.Country.Id, &result.Country.Title, &result.Title, &result.Fake)
		return result, err
	})
	if err != nil {
		return []RegionWithCountry{}, err
	}
	return lst.([]RegionWithCountry), nil
}

func (this regionStorage) Props() PropertyManager {
	return this.PropsManager
}

func (this regionStorage) GetFake(countryId int64) (Region, bool, error) {
	result, found, err := this.doFindAndReturn(this.getFakeQuery, scanFuncI, countryId)
	if err != nil {
		return Region{}, false, err
	}
	if !found {
		return Region{}, false, nil
	}
	return result.(Region), true, nil
}

func (this regionStorage) CreateFake(countryId int64) (int64, error) {
	return this.insertReturningId(this.createFakeQuery, countryId)
}

func scanFuncI(rows *sql.Rows) (interface{}, error) {
	return scanFunc(rows)
}

func scanFunc(rows *sql.Rows) (Region, error) {
	result := Region{}
	err := rows.Scan(&result.Id, &result.CountryId, &result.Title, &result.Fake)
	return result, err
}

