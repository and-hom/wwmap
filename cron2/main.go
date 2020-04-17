package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron2/command"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/gorilla/mux"
	"github.com/robfig/cron"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version string = "development"

func main() {
	log.Infof("Starting wwmap")

	fullConfiguration := config.Load("")
	fullConfiguration.ConfigureLogger()
	configuration := fullConfiguration.Cron

	storage := dao.NewPostgresStorage(fullConfiguration.Db)
	jobDao := NewJobPostgresStorage(storage)
	executionDao := NewExecutionPostgresStorage(storage)
	commands := command.ScanForAvailableCommands()

	logStorage := blob.BasicFsStorage{
		BaseDir: configuration.LogDir,
		Mkdirs:  true,
	}

	registry := CronWithRegistry{
		cron:                     cron.New(),
		jobRegistry:              make(map[int64]cron.EntryID),
		manualRunningJobRegistry: make(map[int64]bool),
		executionDao:             executionDao,
		logStorage:               logStorage,
		commands:                 commands,
	}
	jobs, err := jobDao.list()
	if err != nil {
		log.Fatal("Can't list jobs: ", err)
	}
	for i := 0; i < len(jobs); i++ {
		if err := registry.Register(jobs[i]); err != nil {
			log.Errorf("Can't add job %d: %v", jobs[i].Id, err)
		}
	}
	registry.cron.Start()

	r := mux.NewRouter()

	handler := CronHandler{
		Handler:      Handler{R: r},
		jobDao:       jobDao,
		executionDao: executionDao,
		userDao:      dao.NewUserPostgresDao(storage),
		version:      version,
		enable:       registry.Register,
		disable:      registry.Unregister,
		run:          registry.RunNow,
		commands:     commands,
		commandKeys:  MapKeys(commands),
	}

	handler.Init()

	go func() {
		log.Println("Listening signals...")
		c := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		registry.cron.Stop()

		if err := executionDao.markRunningAsOrphan(); err != nil {
			log.Error("Can't mark running tasks as orphan! ", err)
		}

		os.Exit(0)
	}()

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
