package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func NewCountryPostgresDao(postgresStorage PostgresStorage) CountryDao {
	return countryStorage{
		PostgresStorage: postgresStorage,
		PropsManager:    PropertyManagerImpl{table: queries.SqlQuery("country", "table"), dao: &postgresStorage},
		listQuery:       queries.SqlQuery("country", "list"),
		getQuery:        queries.SqlQuery("country", "get"),
		getByCodeQuery:  queries.SqlQuery("country", "get-by-code"),
		insertQuery:     queries.SqlQuery("country", "insert"),
		updateQuery:     queries.SqlQuery("country", "update"),
		deleteQuery:     queries.SqlQuery("country", "delete"),
	}
}

type countryStorage struct {
	PostgresStorage
	PropsManager   PropertyManager
	listQuery      string
	getQuery       string
	getByCodeQuery string
	insertQuery    string
	updateQuery    string
	deleteQuery    string
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

func (this countryStorage) Save(countries ...Country) error {
	vars := make([]interface{}, len(countries))
	for i, p := range countries {
		vars[i] = p
	}
	return this.PerformUpdates(this.updateQuery, func(entity interface{}) ([]interface{}, error) {
		_country := entity.(Country)
		return []interface{}{_country.Title, _country.Code, _country.Id}, nil
	}, vars...)
}

func (this countryStorage) Insert(country Country) (int64, error) {
	id, err := this.UpdateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(Country)
		return []interface{}{_e.Title, _e.Code}, nil
	}, true, country)
	if err != nil {
		switch e := err.(type) {
		case *pq.Error:
			if e.Code==pq.ErrorCode("23505") {
				return 0, DuplicateError{}
			}
		}
		return 0, err
	}
	return id[0], err
}

func (this countryStorage) Remove(id int64) error {
	log.Infof("Remove country %d", id)
	return this.PerformUpdatesWithinTxOptionally(nil, this.deleteQuery, IdMapper, id)
}

func (this countryStorage) Props() PropertyManager {
	return this.PropsManager
}

func countryMapper(rows *sql.Rows) (Country, error) {
	result := Country{}
	err := rows.Scan(&result.Id, &result.Title, &result.Code)
	return result, err
}
