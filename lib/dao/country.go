package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
	"database/sql"
)

func NewCountryPostgresDao(postgresStorage PostgresStorage) CountryDao {
	return countryStorage{
		PostgresStorage: postgresStorage,
		PropsManager: PropertyManagerImpl{table:queries.SqlQuery("country", "table"), dao:&postgresStorage},
		listQuery: queries.SqlQuery("country", "list"),
	}
}

type countryStorage struct {
	PostgresStorage
	PropsManager PropertyManager
	listQuery    string
}

func (this countryStorage) List() ([]Country, error) {
	lst, err := this.doFindList(this.listQuery, func(rows *sql.Rows) (Country, error) {
		result := Country{}
		err := rows.Scan(&result.Id, &result.Title, &result.Code)
		return result, err
	})
	if err != nil {
		return []Country{}, err
	}
	return lst.([]Country), nil
}

func (this countryStorage) Props() PropertyManager {
	return this.PropsManager
}

