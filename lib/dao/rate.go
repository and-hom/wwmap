package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewCampRatePostgresDao(postgresStorage PostgresStorage) RateDao {
	return NewRatePostgresDao(postgresStorage, "camp_rate")
}

func NewRatePostgresDao(postgresStorage PostgresStorage, table string) RateDao {
	env := make(map[string]string)
	env["table"] = table

	return &rateStorage{
		PostgresStorage: postgresStorage,
		removeQuery:     queries.SqlQueryWithExplicitReplacements("rate", "remove", env),
	}
}

type rateStorage struct {
	PostgresStorage
	removeQuery string
}

func (this rateStorage) RemoveByRefId(refId int64, tx interface{}) error {
	return this.performUpdatesWithinTxOptionally(tx, this.removeQuery, IdMapper, refId)
}
