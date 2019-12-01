package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewCampPhotoPostgresDao(postgresStorage PostgresStorage) PhotoDao {
	return NewPhotoPostgresDao(postgresStorage, "camp_photo")
}

func NewPhotoPostgresDao(postgresStorage PostgresStorage, table string) PhotoDao {
	env := make(map[string]string)
	env["table"] = table

	return &photoStorage{
		PostgresStorage: postgresStorage,
		removeQuery:     queries.SqlQueryWithExplicitReplacements("photo", "remove", env),
	}
}

type photoStorage struct {
	PostgresStorage
	removeQuery string
}

func (this photoStorage) RemoveByRefId(refId int64, tx interface{}) error {
	return this.performUpdatesWithinTxOptionally(tx, this.removeQuery, IdMapper, refId)
}
