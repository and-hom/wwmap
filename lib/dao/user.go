package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"encoding/json"
)

func NewUserPostgresDao(postgresStorage PostgresStorage) UserDao {
	return userStorage{
		PostgresStorage: postgresStorage,
		createQuery: queries.SqlQuery("user", "create"),
		getRoleQuery: queries.SqlQuery("user", "get-role-by-ext-id"),
		setRoleQuery: queries.SqlQuery("user", "set-role"),
		listQuery: queries.SqlQuery("user", "list"),
	}
}

type userStorage struct {
	PostgresStorage
	createQuery  string
	getRoleQuery string
	setRoleQuery string
	listQuery    string
}

func (this userStorage) CreateIfNotExists(user User) error {
	userInfo, err := json.Marshal(user.Info)
	if err != nil {
		return err
	}
	return this.performUpdates(this.createQuery, arrayMapper, []interface{}{user.ExtId, string(user.AuthProvider), user.Role, string(userInfo)})
}

func (this userStorage) List() ([]User, error) {
	result, err := this.doFindList(this.listQuery, func(rows *sql.Rows) (User, error) {
		user := User{}
		infoStr := ""
		authProvider := ""

		err := rows.Scan(&user.Id, &user.ExtId, &authProvider, &user.Role, &infoStr)
		if err != nil {
			return user, err
		}

		err = json.Unmarshal([]byte(infoStr), &user.Info)
		user.AuthProvider = AuthProvider(authProvider)

		return user, err
	})
	if err != nil {
		return []User{}, err
	}
	return result.([]User), nil
}

func (this userStorage) SetRole(userId int64, role Role) error {
	return this.performUpdates(this.setRoleQuery, arrayMapper, []interface{}{userId, role})
}

func (this userStorage) GetRole(provider AuthProvider, extId int64) (Role, error) {
	role := USER
	found, err := this.doFind(this.getRoleQuery, func(rows *sql.Rows) error {
		return rows.Scan(&role)
	}, string(provider), extId)
	if !found {
		return ANONYMOUS, nil
	}
	return role, err
}