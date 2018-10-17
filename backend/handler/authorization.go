package handler

import (
	"net/http"
	. "github.com/and-hom/wwmap/lib/http"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/lib/dao"
	"encoding/json"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/gorilla/mux"
	"strconv"
	"io/ioutil"
	"fmt"
	"time"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/notification"
)

type UserInfoHandler struct {
	App
}

type UserInfoDto struct {
	AuthProvider dao.AuthProvider `json:"auth_provider"`
	Login        string `json:"login"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Roles        []dao.Role `json:"roles"`
}

func (this *UserInfoHandler) Init() {
	this.Register("/user-info", HandlerFunctions{Get: this.GetUserInfo})
	this.Register("/auth-test", HandlerFunctions{Get: this.TestAuth})
	this.Register("/user", HandlerFunctions{Get: this.ListUsers})
	this.Register("/user/{userId}/role", HandlerFunctions{Post: this.SetRole})
	this.Register("/vk/token", HandlerFunctions{Get: this.GetVkToken})
}

func (this *UserInfoHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	authProvider, info, err := this.App.GetUserInfo(r)
	if err != nil {
		onPassportErr(err, w, "Can not get user info from request")
		return
	}

	p, info, err := this.App.GetUserInfo(r)
	if err != nil {
		onPassportErr(err, w, fmt.Sprintf("Can not do request to auth provider: %s", authProvider))
		return
	}

	id, role, justCreated, err := this.CreateMissingUser(r, authProvider, info)
	if err != nil {
		onPassportErr(err, w, "Can not create user!")
		return
	}

	if justCreated {
		this.sendWelcomeMessages(authProvider, id, info)
	}

	infoDto := UserInfoDto{
		FirstName:info.FirstName,
		LastName:info.LastName,
		Login:info.Login,
		Roles:[]dao.Role{role},
		AuthProvider: p,
	}

	bytes, err := json.Marshal(infoDto)
	if err != nil {
		OnError500(w, err, "Can not create response")
		return
	}
	w.Write(bytes)
}

func (this *UserInfoHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN) {
		return
	}

	users, err := this.UserDao.List()
	if err != nil {
		OnError500(w, err, "Can not list users")
		return
	}

	this.JsonAnswer(w, users)
}

func (this *UserInfoHandler) SetRole(w http.ResponseWriter, r *http.Request) {
	if !this.CheckRoleAllowedAndMakeResponse(w, r, dao.ADMIN) {
		return
	}

	pathParams := mux.Vars(r)
	userId, err := strconv.ParseInt(pathParams["userId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		OnError500(w, err, "Can not read body")
		return
	}
	roleStr := ""
	json.Unmarshal(bodyBytes, &roleStr)
	if err != nil {
		OnError500(w, err, "Can not unmarshall role: " + string(bodyBytes))
		return
	}

	if roleStr != string(dao.ADMIN) && roleStr != string(dao.EDITOR) && roleStr != string(dao.USER) {
		OnError(w, err, "Can not set role " + roleStr, http.StatusBadRequest)
		return
	}

	newRole, oldRole, err := this.UserDao.SetRole(userId, dao.Role(roleStr))
	if err != nil {
		OnError500(w, err, "Can not set role")
		return
	}

	users, err := this.UserDao.List()
	if err != nil {
		OnError500(w, err, "Can not list users")
		return
	}

	for i := 0; i < len(users); i++ {
		if users[i].Id == userId && users[i].AuthProvider == dao.YANDEX {
			this.sendChangeRoleMessage(users[i].AuthProvider, users[i].Id, users[i].Info, oldRole, newRole)
		}
	}

	this.JsonAnswer(w, users)
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

type VkTokenAnswer struct {
	AccessToken string `json:"access_token"`
	Expires     int `json:"expires_in"`
	Uid         int64 `json:"uid"`
	ErrDesc     string `json:"error_description"`
}

func (this *UserInfoHandler) GetVkToken(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	resp, err := http.Get("https://oauth.vk.com/access_token?client_id=6703809&client_secret=Q3pUfqqJT77ZbWCyzw5Q&redirect_uri=https://wwmap.ru/redirector-vk.htm&code=" + code)
	if err != nil {
		OnError500(w, err, "Can not get token")
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		OnError500(w, err, "Can not get token")
		return
	}

	answer := VkTokenAnswer{}
	err = json.Unmarshal(body, &answer)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not parse VK response: %s", string(body)))
		return
	}

	if (answer.AccessToken == "") {
		OnError(w, nil, answer.ErrDesc, http.StatusUnauthorized)
	}

	rb, err := json.Marshal(answer.AccessToken)
	if err != nil {
		OnError500(w, err, "Can not marshal response")
		return
	}

	w.Write(rb)
}

func (this *UserInfoHandler) sendChangeRoleMessage(authProvider dao.AuthProvider, userId int64, info passport.UserInfo, oldRole dao.Role, newRole dao.Role) {
	err := this.NotificationDao.Add(dao.Notification{
		Object:dao.IdTitle{Id:userId, Title:info.Login},
		Comment:fmt.Sprintf("%s => %s", oldRole, newRole),
		Recipient:dao.NotificationRecipient{Provider:dao.NOTIFICATION_PROVIDER_EMAIL, Recipient:notification.YandexEmail(info.Login)},
		Classifier:"user-roles",
		SendBefore:time.Now(), // send as soon as possible
	})
	if err != nil {
		log.Errorf("Can not send message to user %d: %v", userId, err)
	}
}

func (this *UserInfoHandler) sendWelcomeMessages(authProvider dao.AuthProvider, id int64, info passport.UserInfo) {
	this.NotificationHelper.SendToRole(dao.Notification{
		IdTitle: dao.IdTitle{Title: string(authProvider)},
		Object: dao.IdTitle{Id:id, Title: fmt.Sprintf("%d %s (%s %s)", info.Id, info.Login, info.FirstName, info.LastName)},
		Comment: "User created",
		Classifier:"user",
		SendBefore:time.Now(), // send as soon as possible
	}, dao.ADMIN)

	if authProvider == dao.YANDEX {
		this.NotificationDao.Add(dao.Notification{
			Object:dao.IdTitle{Id:id, Title:info.Login},
			Recipient:dao.NotificationRecipient{Provider:dao.NOTIFICATION_PROVIDER_EMAIL, Recipient:notification.YandexEmail(info.Login)},
			Classifier:"user-welcome",
			SendBefore:time.Now(), // send as soon as possible
		})
	}
}