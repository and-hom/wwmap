package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
	"database/sql"
)

func NewCountryPostgresDao(postgresStorage PostgresStorage) CountryDao {
	return countryStorage{
		PostgresStorage: postgresStorage,
		listQuery:queries.SqlQuery("country","list"),
	}
}

type countryStorage struct {
	PostgresStorage
	listQuery   string
}

func (this countryStorage) List() ([]Country, error) {
	lst, err := this.doFindList(this.listQuery, func(rows *sql.Rows) (Country,error){
		result := Country{}
		err := rows.Scan(&result.Id, &result.Title)
		return result, err
	})
	if err!=nil {
		return []Country{}, nil
	}
	return lst.([]Country), nil
}

