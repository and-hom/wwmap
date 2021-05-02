package handler

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/notification"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"time"
)

type UserInfoHandler struct {
	App
}

type UserInfoDto struct {
	AuthProvider        dao.AuthProvider `json:"auth_provider"`
	Login               string           `json:"login"`
	FirstName           string           `json:"first_name"`
	LastName            string           `json:"last_name"`
	Roles               []dao.Role       `json:"roles"`
	SessionId           string           `json:"session_id"`
	ExperimentalFeaures bool             `json:"experimental_features"`
}

func (this *UserInfoHandler) Init() {
	this.Register("/user-info", HandlerFunctions{Get: this.GetUserInfo})
	this.Register("/session-start", HandlerFunctions{Get: this.SessionStart})
	this.Register("/auth-test", HandlerFunctions{Get: this.TestAuth})
	this.Register("/user", HandlerFunctions{Get: this.ForRoles(this.ListUsers, dao.ADMIN)})
	this.Register("/user/{userId}/role", HandlerFunctions{Post: this.ForRoles(this.SetRole, dao.ADMIN)})
	this.Register("/user/{userId}/experimental", HandlerFunctions{Post: this.ForRoles(this.SetExperimentalFeatures, dao.ADMIN)})
	this.Register("/vk/token", HandlerFunctions{Get: this.GetVkToken})
}

func (this *UserInfoHandler) SessionStart(w http.ResponseWriter, r *http.Request) {
	authProvider, info, err := this.App.GetUserInfo(r)
	if err != nil {
		onPassportErr(err, w, fmt.Sprintf("Can not do request to auth provider: %s", authProvider))
		return
	}

	newSessionId := uuid.New().String()

	id, role, sessionId, justCreated, err := this.CreateMissingUser(r, authProvider, info, newSessionId)
	if err != nil {
		onPassportErr(err, w, "Can not create user!")
		return
	}

	if justCreated {
		this.sendWelcomeMessages(authProvider, id, info)

		rWithUser := r.WithContext(context.WithValue(r.Context(), USER_REQUEST_VARIABLE, &dao.User{
			ExtId:        info.Id,
			AuthProvider: authProvider,
			Info:         dao.UserInfo{Login: info.Login},
		}))
		this.LogUserEvent(rWithUser, USER_LOG_ENTRY_TYPE, id, dao.ENTRY_TYPE_CREATE,
			fmt.Sprintf("%s (%s %s)", info.Login, info.FirstName, info.LastName))
	}

	infoDto := UserInfoDto{
		FirstName:           info.FirstName,
		LastName:            info.LastName,
		Login:               info.Login,
		Roles:               []dao.Role{role},
		AuthProvider:        authProvider,
		SessionId:           sessionId,
		ExperimentalFeaures: false,
	}

	JsonAnswer(w, infoDto)
}

func (this *UserInfoHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	sessionId := r.FormValue("session_id")
	user, err := this.UserDao.GetBySession(sessionId)
	if err != nil {
		onPassportErr(err, w, "Can not get user info!")
		return
	}

	infoDto := UserInfoDto{
		FirstName:           user.Info.FirstName,
		LastName:            user.Info.LastName,
		Login:               user.Info.Login,
		Roles:               []dao.Role{user.Role},
		AuthProvider:        user.AuthProvider,
		SessionId:           user.SessionId,
		ExperimentalFeaures: user.ExperimentalFeaures,
	}

	JsonAnswer(w, infoDto)
}

func (this *UserInfoHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := this.UserDao.List()
	if err != nil {
		OnError500(w, err, "Can not list users")
		return
	}

	JsonAnswer(w, users)
}

func (this *UserInfoHandler) SetRole(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	userId, err := strconv.ParseInt(pathParams["userId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	roleStr := ""
	body, err := DecodeJsonBody(r, &roleStr)
	if err != nil {
		OnError500(w, err, "Can not unmarshall role: "+body)
		return
	}

	if roleStr != string(dao.ADMIN) && roleStr != string(dao.EDITOR) && roleStr != string(dao.USER) {
		OnError(w, err, "Can not set role "+roleStr, http.StatusBadRequest)
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

	user := dao.User{}
	for i := 0; i < len(users); i++ {
		if users[i].Id == userId {
			this.sendChangeRoleMessage(users[i].Id, users[i].Info, users[i].AuthProvider, oldRole, newRole)
		}
		if users[i].Id == userId {
			user = users[i]
		}
	}

	JsonAnswer(w, users)

	loginPrefix := ""
	if user.AuthProvider == dao.VK {
		loginPrefix = string(user.AuthProvider) + "/"
	}

	this.LogUserEvent(r, USER_LOG_ENTRY_TYPE, userId, dao.ENTRY_TYPE_MODIFY, fmt.Sprintf("%s%s (%s %s) %s => %s",
		loginPrefix, user.Info.Login, user.Info.FirstName, user.Info.LastName, oldRole, newRole))
}

func (this *UserInfoHandler) SetExperimentalFeatures(w http.ResponseWriter, r *http.Request) {
	pathParams := mux.Vars(r)
	userId, err := strconv.ParseInt(pathParams["userId"], 10, 64)
	if err != nil {
		OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	experimentalFeatures := false
	body, err := DecodeJsonBody(r, &experimentalFeatures)
	if err != nil {
		OnError500(w, err, "Can not unmarshall request body: "+body)
		return
	}

	newExperimentalMode, oldExperimentalMode, err := this.UserDao.SetExperimentalFeatures(userId, experimentalFeatures)
	if err != nil {
		OnError500(w, err, "Can not set experimental features mode")
		return
	}

	users, err := this.UserDao.List()
	if err != nil {
		OnError500(w, err, "Can not list users")
		return
	}

	for i := 0; i < len(users); i++ {
		if users[i].Id == userId {
			this.sendChangeExperimentalModeMessage(users[i].AuthProvider, users[i].Id, users[i].Info, newExperimentalMode)
		}
	}

	JsonAnswer(w, users)

	this.LogUserEvent(r, USER_LOG_ENTRY_TYPE, userId, dao.ENTRY_TYPE_MODIFY, fmt.Sprintf("Exp. %t => %t", oldExperimentalMode, newExperimentalMode))
}

func (this *UserInfoHandler) TestAuth(w http.ResponseWriter, r *http.Request) {
	_, found, err := CheckRoleAllowed(r, this.UserDao, dao.ADMIN)
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
	Expires     int    `json:"expires_in"`
	Uid         int64  `json:"uid"`
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

	answer := VkTokenAnswer{}
	err = json.NewDecoder(resp.Body).Decode(&answer)
	if err != nil {
		OnError500(w, err, "Can not parse VK response")
		return
	}

	if answer.AccessToken == "" {
		OnError(w, nil, answer.ErrDesc, http.StatusUnauthorized)
	}

	JsonAnswer(w, answer)
}

func (this *UserInfoHandler) sendChangeRoleMessage(userId int64, info dao.UserInfo, authProvider dao.AuthProvider, oldRole dao.Role, newRole dao.Role) {
	recipient := notification.GetRecipient(authProvider, info)
	if recipient == nil {
		return
	}

	err := this.NotificationDao.Add(dao.Notification{
		Object:     dao.IdTitle{Id: userId, Title: info.Login},
		Comment:    fmt.Sprintf("%s => %s", oldRole, newRole),
		Recipient:  *recipient,
		Classifier: "user-roles",
		SendBefore: time.Now(), // send as soon as possible
	})
	if err != nil {
		log.Errorf("Can not send message to user %d: %v", userId, err)
	}
}

func (this *UserInfoHandler) sendChangeExperimentalModeMessage(authProvider dao.AuthProvider, userId int64, info dao.UserInfo, new bool) {
	recipient := notification.GetRecipient(authProvider, info)
	if recipient == nil {
		return
	}

	comment := ""
	if new {
		comment = "включены"
	} else {
		comment = "выключены"
	}
	err := this.NotificationDao.Add(dao.Notification{
		Object:     dao.IdTitle{Id: userId, Title: info.Login},
		Comment:    comment,
		Recipient:  *recipient,
		Classifier: "user-experimental-features",
		SendBefore: time.Now(), // send as soon as possible
	})
	if err != nil {
		log.Errorf("Can not send message to user %d: %v", userId, err)
	}
}

func (this *UserInfoHandler) sendWelcomeMessages(authProvider dao.AuthProvider, id int64, info passport.UserInfo) {
	if err := this.NotificationHelper.SendToRole(dao.Notification{
		IdTitle:    dao.IdTitle{Title: string(authProvider)},
		Object:     dao.IdTitle{Id: id, Title: fmt.Sprintf("%s %s (%s %s)", info.Id, info.Login, info.FirstName, info.LastName)},
		Comment:    "User created",
		Classifier: "user",
		SendBefore: time.Now(), // send as soon as possible
	}, dao.ADMIN); err != nil {
		log.Errorf("Can't create new user notification for admin: %v", err)
	}

	recipient := notification.GetRecipient(authProvider, Convert(info))
	if recipient == nil {
		return
	}

	if err := this.NotificationDao.Add(dao.Notification{
		IdTitle:    dao.IdTitle{Title: string(authProvider)},
		Object:     dao.IdTitle{Id: id, Title: string(info.Login)},
		Recipient:  *recipient,
		Classifier: "user-welcome",
		SendBefore: time.Now(), // send as soon as possible
	}); err != nil {
		log.Errorf("Can't create welcome notification for user: %v", err)
	}
}

func Convert(info passport.UserInfo) dao.UserInfo {
	return dao.UserInfo{
		Login:     info.Login,
		FirstName: info.FirstName,
		LastName:  info.LastName,
	}
}
