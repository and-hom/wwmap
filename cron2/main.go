package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

var version string = "development"

func main() {
	log.Infof("Starting wwmap")

	fullConfiguration := config.Load("")
	fullConfiguration.ChangeLogLevel()
	configuration := fullConfiguration.Cron

	storage := dao.NewPostgresStorage(fullConfiguration.Db)

	r := mux.NewRouter()

	handler := CronHandler{
		Handler:      Handler{R: r},
		jobDao:       NewJobPostgresStorage(storage),
		executionDao: NewExecutionPostgresStorage(storage),
		userDao:      dao.NewUserPostgresDao(storage),
		version:      version,
	}

	handler.Init()

	log.Infof("Starting tiles server on %v+", configuration.BindTo)

	srv := &http.Server{
		ReadTimeout: 5 * time.Second,
		Addr:        configuration.BindTo,
		Handler:     WrapWithLogging(r, fullConfiguration),
	}
	if configuration.ReadTimeout > 0 {
		srv.ReadTimeout = configuration.ReadTimeout
	}
	if configuration.WriteTimeout > 0 {
		srv.WriteTimeout = configuration.WriteTimeout
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
