package handler

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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
