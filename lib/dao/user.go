package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"encoding/json"
	"github.com/pkg/errors"
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

func (this userStorage) CreateIfNotExists(user User) (int64, Role, bool, error) {
	userInfo, err := json.Marshal(user.Info)
	if err != nil {
		return 0, ANONYMOUS, false, err
	}
	cols, err := this.updateReturningColumns(this.createQuery, arrayMapper, []interface{}{user.ExtId, string(user.AuthProvider), user.Role, string(userInfo)})
	if err != nil {
		return 0, ANONYMOUS, false, err
	}
	if len(cols) < 1 {
		return 0, ANONYMOUS, false, errors.New("User id and created flag were not returned! Empty row!")
	}
	if len(cols[0]) < 3 {
		return 0, ANONYMOUS, false, errors.New("User id and created flag were not returned! Too short row, should be length==3")
	}
	return *(cols[0][0].(*int64)), Role(*(cols[0][1].(*string))), *(cols[0][2].(*bool)), nil
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