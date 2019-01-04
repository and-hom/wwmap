package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"encoding/json"
	"github.com/pkg/errors"
	"fmt"
)

func NewUserPostgresDao(postgresStorage PostgresStorage) UserDao {
	return userStorage{
		PostgresStorage: postgresStorage,
		createQuery: queries.SqlQuery("user", "create"),
		getRoleQuery: queries.SqlQuery("user", "get-role-by-ext-id"),
		setRoleQuery: queries.SqlQuery("user", "set-role"),
		listQuery: queries.SqlQuery("user", "list"),
		listByRoleQuery: queries.SqlQuery("user", "list-by-role"),
		getBySessionQuery: queries.SqlQuery("user", "find-by-session-id"),
	}
}

type userStorage struct {
	PostgresStorage
	createQuery     string
	getRoleQuery    string
	setRoleQuery    string
	listQuery       string
	listByRoleQuery string
	getBySessionQuery string
}

func (this userStorage) CreateIfNotExists(user User) (int64, Role, string, bool, error) {
	userInfo, err := json.Marshal(user.Info)
	if err != nil {
		return 0, ANONYMOUS, user.SessionId, false, err
	}
	cols, err := this.updateReturningColumns(this.createQuery, ArrayMapper, []interface{}{user.ExtId, string(user.AuthProvider), user.Role, string(userInfo), user.SessionId})
	if err != nil {
		return 0, ANONYMOUS, user.SessionId, false, err
	}
	if len(cols) < 1 {
		return 0, ANONYMOUS, user.SessionId, false, errors.New("User id and created flag were not returned! Empty row!")
	}
	if len(cols[0]) < 4 {
		return 0, ANONYMOUS, user.SessionId, false, errors.New("User id and created flag were not returned! Too short row, should be length==3")
	}
	return *(cols[0][0].(*int64)), Role(*(cols[0][1].(*string))), *(cols[0][2].(*string)), *(cols[0][3].(*bool)), nil
}

func (this userStorage) List() ([]User, error) {
	result, err := this.doFindList(this.listQuery, userMapper)
	if err != nil {
		return []User{}, err
	}
	return result.([]User), nil
}
func (this userStorage) ListByRole(role Role) ([]User, error) {
	result, err := this.doFindList(this.listByRoleQuery, userMapper, role)
	if err != nil {
		return []User{}, err
	}
	return result.([]User), nil
}

func (this userStorage) SetRole(userId int64, role Role) (Role, Role, error) {
	cols, err := this.updateReturningColumns(this.setRoleQuery, ArrayMapper, []interface{}{userId, role})
	if err != nil {
		return ANONYMOUS, ANONYMOUS, err
	}
	if len(cols) == 0 {
		return ANONYMOUS, ANONYMOUS, fmt.Errorf("Can not update role for user %d: user not found", userId)
	}
	return Role(*(cols[0][0].(*string))), Role(*(cols[0][1].(*string))), nil
}

func (this userStorage) GetBySession(sessionId string) (User, error) {
	user, found, err := this.doFindAndReturn(this.getBySessionQuery, userMapper, sessionId)
	if err!=nil {
		return User{}, err
	}
	if !found {
		return User{}, passport.UnauthorizedError{}
	}
	return user.(User), nil
}

func (this userStorage) GetRole(provider AuthProvider, extId string) (Role, error) {
	role, found, err := this.doFindAndReturn(this.getRoleQuery, func(rows *sql.Rows) error {
		role := USER
		return rows.Scan(&role)
	}, string(provider), extId)
	if err!=nil {
		return ANONYMOUS, err
	}
	if !found {
		return ANONYMOUS, nil
	}
	return role.(Role), err
}

func userMapper(rows *sql.Rows) (User, error) {
	user := User{}
	infoStr := ""
	authProvider := ""
	sessionId := sql.NullString{}

	err := rows.Scan(&user.Id, &user.ExtId, &authProvider, &user.Role, &infoStr, &sessionId)
	if err != nil {
		return user, err
	}
	if sessionId.Valid {
		user.SessionId = sessionId.String
	}

	err = json.Unmarshal([]byte(infoStr), &user.Info)
	user.AuthProvider = AuthProvider(authProvider)

	return user, err
}
