package handler

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	OPTIONS = "OPTIONS"
	HEAD    = "HEAD"
	GET     = "GET"
	PUT     = "PUT"
	POST    = "POST"
	DELETE  = "DELETE"
)

const USER_REQUEST_VARIABLE = "user"

type ApiHandler interface {
	Init()
}

type HandlerFunction func(http.ResponseWriter, *http.Request)

type HandlerFunctions struct {
	Head   HandlerFunction
	Get    HandlerFunction
	Post   HandlerFunction
	Put    HandlerFunction
	Delete HandlerFunction
}

func (this *HandlerFunctions) CorsMethods() []string {
	corsMethods := []string{}
	if this.Head != nil {
		corsMethods = append(corsMethods, HEAD)
	}
	if this.Get != nil {
		corsMethods = append(corsMethods, GET)
	}
	if this.Post != nil {
		corsMethods = append(corsMethods, POST)
	}
	if this.Put != nil {
		corsMethods = append(corsMethods, PUT)
	}
	if this.Delete != nil {
		corsMethods = append(corsMethods, DELETE)
	}
	return corsMethods
}

type Handler struct {
	R *mux.Router
}

func (this *Handler) CorOptionsStub(w http.ResponseWriter, r *http.Request, corsMethods []string) {
	CorsHeaders(w, corsMethods...)
	// for cors only
}

func (this *Handler) Register(path string, handlerFunctions HandlerFunctions) {
	corsMethods := handlerFunctions.CorsMethods()

	this.R.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		this.CorOptionsStub(w, r, corsMethods)
	}).Methods(OPTIONS)

	this.registerOne(path, GET, handlerFunctions.Get, corsMethods)
	this.registerOne(path, HEAD, handlerFunctions.Head, corsMethods)
	this.registerOne(path, PUT, handlerFunctions.Put, corsMethods)
	this.registerOne(path, POST, handlerFunctions.Post, corsMethods)
	this.registerOne(path, DELETE, handlerFunctions.Delete, corsMethods)
}

func (this *Handler) registerOne(path string, method string, handlerFunction HandlerFunction, corsMethods []string) {
	if handlerFunction == nil {
		return
	}
	this.R.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		CorsHeaders(w, corsMethods...)
		handlerFunction(w, r)
	}).Methods(method)
}

func (this *Handler) JsonpAnswer(callback string, object interface{}, _default string) []byte {
	return []byte(callback + "(" + this.JsonStr(object, _default) + ");")
}

func (this *Handler) JsonStr(f interface{}, _default string) string {
	bytes, err := json.Marshal(f)
	if err != nil {
		log.Errorf("Can not serialize object %v: %s", f, err.Error())
		return _default
	}
	return string(bytes)
}

func (this *Handler) JsonAnswerF(w http.ResponseWriter, f func() (interface{}, error), errStr string) {
	payload, err := f()
	if err != nil {
		OnError500(w, err, errStr)
	} else {
		this.JsonAnswer(w, payload)
	}
}

func (this *Handler) JsonAnswer(w http.ResponseWriter, f interface{}) {
	bytes, err := json.Marshal(f)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize object %v", f))
		return
	}

	SetJsonResponseHeaders(w)
	w.Write(bytes)
}

func WrapWithLogging(h http.Handler, configuration config.Configuration) http.Handler {
	logLevel, err := configuration.LogLevel.ToLogrus()
	if err != nil {
		log.Fatalf("Can not parse log level %s", configuration.LogLevel)
	}
	if logLevel == log.DebugLevel {
		return handlers.LoggingHandler(os.Stdout, h)
	}
	return h
}

func DecodeJsonBody(r *http.Request, obj interface{}) (string, error) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bodyBytes, obj)
	return string(bodyBytes), err
}

func CheckRoleAllowed(r *http.Request, userDao dao.UserDao, allowedRoles ...dao.Role) (*http.Request, bool, error) {
	sessionId := r.FormValue("session_id")
	if sessionId == "" {
		sessionId = r.Header.Get("Authorization")
	}
	if sessionId == "" {
		return r, false, nil
	}
	user, err := userDao.GetBySession(sessionId)
	if err != nil {
		return r, false, err
	}
	rWithUser := r.WithContext(context.WithValue(r.Context(), USER_REQUEST_VARIABLE, &user))

	for i := 0; i < len(allowedRoles); i++ {
		if allowedRoles[i] == user.Role {
			return rWithUser, true, nil
		}
	}
	return rWithUser, false, nil
}

func ForRoles(payload HandlerFunction, userDao dao.UserDao, roles ...dao.Role) HandlerFunction {
	if len(roles) == 0 {
		return payload
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		r2, ok := CheckRoleAllowedAndMakeResponse(writer, userDao, request, roles...)
		if ok {
			payload(writer, r2)
		}
	}
}

func CheckRoleAllowedAndMakeResponse(w http.ResponseWriter, userDao dao.UserDao, r *http.Request, allowedRoles ...dao.Role) (*http.Request, bool) {
	r2, allowed, err := CheckRoleAllowed(r, userDao, allowedRoles...)
	if err != nil {
		switch err.(type) {
		default:
			OnError500(w, err, "Can not check permissions")
		case passport.UnauthorizedError:
			OnError(w, err, "Unauthorized", http.StatusUnauthorized)
		}
		return r2, false
	}
	if !allowed {
		msg := ""
		if len(allowedRoles) == 1 {
			msg = fmt.Sprintf("Sorry! You haven't role %s", allowedRoles[0])
		} else {
			msg = fmt.Sprintf("Sorry! You haven't any of following roles: %s", dao.Join(", ", allowedRoles...))
		}
		OnError(w, nil, msg, http.StatusUnauthorized)
		return r2, false
	}
	return r2, true
}

func ShowUnpublished(req *http.Request, userDao dao.UserDao) bool {
	showUnpublishedStr := req.FormValue("show_unpublished")
	showUnpublished := false
	if showUnpublishedStr == "true" || showUnpublishedStr == "1" {
		_, allowed, err := CheckRoleAllowed(req, userDao, dao.ADMIN, dao.EDITOR)
		showUnpublished = err == nil && allowed
	}
	return showUnpublished
}
