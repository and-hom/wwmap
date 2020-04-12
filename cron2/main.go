package main

import (
	"fmt"
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
	fullConfiguration.ChangeLogLevel()
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
		cron:         cron.New(),
		jobRegistry:  make(map[int64]cron.EntryID),
		executionDao: executionDao,
		logStorage:   logStorage,
		commands:     commands,
	}
	jobs, err := jobDao.list()
	if err != nil {
		log.Fatal("Can't list jobs: ", err)
	}
	for i := 0; i < len(jobs); i++ {
		if err := registry.Register(jobs[i]); err != nil {
			log.Fatalf("Can't add job %d: %v", jobs[i], err)
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

type CronWithRegistry struct {
	cron         *cron.Cron
	jobRegistry  map[int64]cron.EntryID
	logStorage   blob.BlobStorage
	executionDao ExecutionDao
	commands     map[string]command.Command
}

func (this CronWithRegistry) Register(job Job) error {
	if !job.Enabled {
		log.Infof("Skip diabled job %d %s %s \"%s\"", job.Id, job.Title, job.Expr, job.Command)
		return nil
	}

	c, commandFound := this.commands[job.Command]
	if !commandFound {
		log.Warnf("Skip job %d with missing command %s", job.Id, job.Command)
		return nil
	}

	log.Infof("Register job %d %s %s \"%s\"", job.Id, job.Title, job.Expr, job.Command)
	runner := Runner{
		Job:          job,
		Command:      c,
		BlobStorage:  this.logStorage,
		ExecutionDao: this.executionDao,
	}
	entryId, err := this.cron.AddFunc(job.Expr, runner.Run)
	if err != nil {
		return err
	}

	this.jobRegistry[job.Id] = entryId
	return nil
}

func (this CronWithRegistry) Unregister(jobId int64) error {
	entryId, ok := this.jobRegistry[jobId]
	if !ok {
		return fmt.Errorf("Job with id=%d is not registered in cron", jobId)
	}
	this.cron.Remove(entryId)
	delete(this.jobRegistry, jobId)
	return nil
}

func MapKeys(m map[string]command.Command) []string {
	result := make([]string, 0, len(m))
	for k, _ := range m {
		result = append(result, k)
	}
	return result
}
