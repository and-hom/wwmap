package handler

import (
	"net/http"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/lib/dao"
	"encoding/json"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/gorilla/mux"
)

type UserInfoHandler struct {
	App
}

type UserInfoDto struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Roles     []dao.Role `json:"roles"`
}

func (this *UserInfoHandler) Init(r *mux.Router) {
	this.Register(r, "/user-info", HandlerFunctions{Get: this.GetUserInfo})
	this.Register(r, "/auth-test", HandlerFunctions{Get: this.TestAuth})
}

func (this *UserInfoHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	err := this.CreateMissingUser(r)
	if err != nil {
		onPassportErr(err, w, "Can not create user")
		return
	}

	token := GetOauthToken(r)
	info, err := this.YandexPassport.ResolveUserInfo(token)
	if err != nil {
		onPassportErr(err, w, "Can not do request to Yandex Passport")
		return
	}

	role, err := this.UserDao.GetRole(info.Id)
	if err != nil {
		onPassportErr(err, w, "Can not get role for user")
		return
	}

	infoDto := UserInfoDto{
		FirstName:info.FirstName,
		LastName:info.LastName,
		Roles:[]dao.Role{role},
	}

	bytes, err := json.Marshal(infoDto)
	if err != nil {
		OnError500(w, err, "Can not create response")
		return
	}
	w.Write(bytes)
}

func (this *UserInfoHandler) TestAuth(w http.ResponseWriter, r *http.Request) {
	found, err := this.CheckRoleAllowed(r, dao.ADMIN)
	if err != nil {
		onPassportErr(err, w, "Can not do request to Yandex Passport")
		return
	}
	if !found {
		OnError(w, nil, "User not found", http.StatusUnauthorized)
	}
}

func onPassportErr(err error, w http.ResponseWriter, msg string) {
	switch err.(type) {
	case passport.UnauthorizedError:
		OnError(w, nil, "User not found", http.StatusUnauthorized)
	default:
		OnError500(w, err, msg)
	}
}