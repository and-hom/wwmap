package http

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"net/http"
	"strings"
)

func CorsHeaders(w http.ResponseWriter, methods ...string) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", strings.Join(methods, ","))
	w.Header().Add("Access-Control-Allow-Headers", "origin, x-csrftoken, content-type, accept, authorization")
}

func SetJsonResponseHeaders(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
}

func OnErrorWithCustomLogging(w http.ResponseWriter, err error, msg string, statusCode int, logF func(string)) {
	errStr := msg
	if err != nil {
		errStr = fmt.Sprintf("%s: %v", msg, err)
	}
	logF(errStr)
	http.Error(w, errStr, statusCode)
}

func OnError(w http.ResponseWriter, err error, msg string, statusCode int) {
	OnErrorWithCustomLogging(w, err, msg, statusCode, func(msg string) {
		log.Error(msg)
	})
}

func OnError500(w http.ResponseWriter, err error, msg string) {
	OnError(w, err, msg, http.StatusInternalServerError)
}

type AuthProviderAndToken struct {
	AuthProvider dao.AuthProvider
	Token        string
}

func GetOauthProviderAndToken(r *http.Request) AuthProviderAndToken {
	var token string
	var provider string

	authorization := r.Header.Get("Authorization")
	parts := strings.Split(authorization, " ")
	if len(parts) > 1 && parts[1] != "" {
		provider = parts[0]
		token = parts[1]
	} else {
		provider = r.FormValue("provider")
		token = r.FormValue("token")
	}

	return AuthProviderAndToken{
		AuthProvider: dao.AuthProvider(provider),
		Token:        token,
	}
}
