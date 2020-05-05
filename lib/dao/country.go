package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewCountryPostgresDao(postgresStorage PostgresStorage) CountryDao {
	return countryStorage{
		PostgresStorage: postgresStorage,
		PropsManager:    PropertyManagerImpl{table: queries.SqlQuery("country", "table"), dao: &postgresStorage},
		listQuery:       queries.SqlQuery("country", "list"),
		getQuery:        queries.SqlQuery("country", "get"),
		getByCodeQuery:  queries.SqlQuery("country", "get-by-code"),
	}
}

type countryStorage struct {
	PostgresStorage
	PropsManager   PropertyManager
	listQuery      string
	getQuery       string
	getByCodeQuery string
}

func (this countryStorage) List() ([]Country, error) {
	lst, err := this.DoFindList(this.listQuery, countryMapper)
	if err != nil {
		return []Country{}, err
	}
	return lst.([]Country), nil
}

func (this countryStorage) Get(id int64) (Country, bool, error) {
	p, found, err := this.DoFindAndReturn(this.getQuery, countryMapper, id)
	if err != nil {
		return Country{}, false, err
	}
	if !found {
		return Country{}, false, nil
	}
	return p.(Country), true, nil
}

func (this countryStorage) GetByCode(code string) (Country, bool, error) {
	p, found, err := this.DoFindAndReturn(this.getByCodeQuery, countryMapper, code)
	if err != nil {
		return Country{}, false, err
	}
	if !found {
		return Country{}, false, nil
	}
	return p.(Country), true, nil
}

func (this countryStorage) Props() PropertyManager {
	return this.PropsManager
}

func countryMapper(rows *sql.Rows) (Country, error) {
	result := Country{}
	err := rows.Scan(&result.Id, &result.Title, &result.Code)
	return result, err
}
