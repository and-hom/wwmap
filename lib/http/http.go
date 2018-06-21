package http

import (
	"net/http"
	"fmt"
	log "github.com/Sirupsen/logrus"
)

func CorsHeaders(w http.ResponseWriter, methods string) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", methods)
	w.Header().Add("Access-Control-Allow-Headers", "origin, x-csrftoken, content-type, accept")
}

func JsonResponse(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
}

func OnError(w http.ResponseWriter, err error, msg string, statusCode int) {
	errStr := fmt.Sprintf("%s: %v", msg, err)
	log.Errorf(errStr)
	http.Error(w, errStr, statusCode)
}

func OnError500(w http.ResponseWriter, err error, msg string) {
	OnError(w, err, msg, http.StatusInternalServerError)
}

func GetOauthToken(r *http.Request) string {
	return r.FormValue("token")
}