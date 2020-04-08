package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/gorilla/mux"
	"github.com/robfig/cron"
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
	jobDao := NewJobPostgresStorage(storage)
	executionDao := NewExecutionPostgresStorage(storage)

	logStorage := blob.BasicFsStorage{BaseDir: configuration.LogDir}

	c := cron.New()
	jobs, err := jobDao.list()
	if err != nil {
		log.Fatal("Can't list jobs: ", err)
	}
	for i := 0; i < len(jobs); i++ {
		if !jobs[i].Enabled {
			log.Infof("Skip diabled job %d %s %s \"%s\"", jobs[i].Id, jobs[i].Title, jobs[i].Expr, jobs[i].Command)
			continue
		}
		log.Infof("Register job %d %s %s \"%s\"", jobs[i].Id, jobs[i].Title, jobs[i].Expr, jobs[i].Command)
		runner := Runner{
			Job:          jobs[i],
			BlobStorage:  logStorage,
			ExecutionDao: executionDao,
		}
		_, err = c.AddFunc(jobs[i].Expr, runner.Run)
		if err != nil {
			log.Fatalf("Can't add job %d: %v", jobs[i], err)
		}
	}
	c.Start()

	r := mux.NewRouter()

	handler := CronHandler{
		Handler:      Handler{R: r},
		jobDao:       jobDao,
		executionDao: executionDao,
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
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
