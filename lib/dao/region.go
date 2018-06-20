package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
	"database/sql"
)

func NewRegionPostgresDao(postgresStorage PostgresStorage) RegionDao {
	return regionStorage{
		PostgresStorage: postgresStorage,
		listQuery:queries.SqlQuery("region", "list-real"),
	}
}

type regionStorage struct {
	PostgresStorage
	listQuery string
}

func (this regionStorage) List(countryId int64) ([]Region, error) {
	lst, err := this.doFindList(this.listQuery, func(rows *sql.Rows) (Region, error) {
		result := Region{}
		err := rows.Scan(&result.Id, &result.CountryId, &result.Title)
		return result, err
	}, countryId)
	if err != nil {
		return []Region{}, nil
	}
	return lst.([]Region), nil
}

