package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewUserPostgresDao(postgresStorage PostgresStorage) UserDao {
	return userStorage{
		PostgresStorage: postgresStorage,
		createQuery: queries.SqlQuery("user", "create"),
		getRoleQuery: queries.SqlQuery("user", "get-role-by-yandex-id"),
	}
}

type userStorage struct {
	PostgresStorage
	createQuery  string
	getRoleQuery string
}

func (this userStorage) CreateIfNotExists(user User) error {
	_, err := this.insertReturningId(this.createQuery, user.YandexId, user.Role, "{}")
	return err
}

func (this userStorage) GetRole(yandexId int64) (Role, error) {
	role := USER
	found, err := this.doFind(this.getRoleQuery, func(rows *sql.Rows) error {
		return rows.Scan(&role)
	}, yandexId)
	if !found {
		return ANONYMOUS, nil
	}
	return role, err
}