package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron2/command"
	cronDao "github.com/and-hom/wwmap/cron2/dao"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/handler"
	http2 "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const DEFAULT_TIMELINE_DURATION = 24 * time.Hour
const MAX_LOG_SIZE = 20 * 1024 * 1024

type CronHandler struct {
	handler.Handler
	jobDao       cronDao.JobDao
	executionDao cronDao.ExecutionDao
	userDao      dao.UserDao
	logStorage   blob.BlobStorage
	registry     CronWithRegistry
	version      string
	enable       func(cronDao.Job) error
	disable      func(int64) error
	run          func(cronDao.Job) error
	commands     map[string]command.Command
	commandKeys  []string
}

func (this *CronHandler) Init() {
	this.Register("/commands", handler.HandlerFunctions{
		Get: this.ForRoles(this.Commands, dao.ADMIN),
	})
	this.Register("/job", handler.HandlerFunctions{
		Get:  this.ForRoles(this.List, dao.ADMIN),
		Put:  this.ForRoles(this.Upsert, dao.ADMIN),
		Post: this.ForRoles(this.Upsert, dao.ADMIN),
	})
	this.Register("/job/{id}", handler.HandlerFunctions{
		Get:    this.ForRoles(this.Get, dao.ADMIN),
		Put:    this.ForRoles(this.Upsert, dao.ADMIN),
		Post:   this.ForRoles(this.Upsert, dao.ADMIN),
		Delete: this.ForRoles(this.Delete, dao.ADMIN),
	})
	this.Register("/job/{id}/run", handler.HandlerFunctions{
		Post: this.ForRoles(this.Run, dao.ADMIN),
	})
	this.Register("/timeline", handler.HandlerFunctions{Get: this.ForRoles(this.Timeline, dao.ADMIN)})
	this.Register("/logs/{id}/{qualifier}", handler.HandlerFunctions{Get: this.ForRoles(this.Logs, dao.ADMIN)})
	this.Register("/version", handler.HandlerFunctions{Get: this.Version})
	this.Register("/health", handler.HandlerFunctions{Get: this.Health})
}

func (this *CronHandler) ForRoles(payload handler.HandlerFunction, roles ...dao.Role) handler.HandlerFunction {
	return handler.ForRoles(payload, this.userDao, roles...)
}

func (this *CronHandler) Commands(w http.ResponseWriter, req *http.Request) {
	this.JsonAnswer(w, this.commands)
}

func (this *CronHandler) List(w http.ResponseWriter, req *http.Request) {
	jobs, err := this.jobDao.List()
	if err != nil {
		http2.OnError500(w, err, "Can't list jobs")
		return
	}

	out := make([]JobDto, len(jobs))
	for i := 0; i < len(jobs); i++ {
		_, registered := this.registry.jobRegistry[jobs[i].Id]
		unregisteredReason, _ := this.registry.unregisteredReasons[jobs[i].Id]
		out[i] = JobDto{jobs[i], registered, unregisteredReason}
	}
	this.JsonAnswer(w, out)
}

func (this *CronHandler) Get(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		http2.OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	jobs, found, err := this.jobDao.Get(id)
	if err != nil {
		http2.OnError500(w, err, fmt.Sprintf("Can't get job with id %d", id))
		return
	}
	if !found {
		http2.OnError(w, err, fmt.Sprintf("Job with id %d not exists", id), http.StatusNotFound)
		return
	}
	this.JsonAnswer(w, jobs)
}

func (this *CronHandler) Upsert(w http.ResponseWriter, req *http.Request) {
	job := cronDao.Job{}
	body, err := handler.DecodeJsonBody(req, &job)
	if err != nil {
		http2.OnError500(w, err, "Can not parse json from request body: "+body)
		return
	}

	if _, ok := this.commands[job.Command]; !ok {
		http2.OnError(w, fmt.Errorf("Invalid command %s", job.Command), "Illegal job spec:", http.StatusBadRequest)
		return
	}

	if job.Id > 0 {
		var enabledStateChanged bool
		enabledStateChanged, err = this.jobDao.Update(job)
		if err != nil {
			http2.OnError500(w, err, "Can't update")
			return
		}
		if this.changeJobState(enabledStateChanged, job, w) {
			return
		}
	} else {
		id, err := this.jobDao.Insert(job)
		if err != nil {
			http2.OnError500(w, err, "Can't insert")
			return
		}
		job.Id = id
		if this.changeJobState(job.Enabled, job, w) {
			return
		}
	}
	this.JsonAnswer(w, true)
}

func (this *CronHandler) changeJobState(enabledStateChanged bool, job cronDao.Job, w http.ResponseWriter) bool {
	if enabledStateChanged {
		_, registered := this.registry.jobRegistry[job.Id]
		if registered {
			if err := this.disable(job.Id); err != nil {
				http2.OnError500(w, err, "Can't unregister job from cron!")
				return true
			}
		}
		if job.Enabled {
			if err := this.enable(job); err != nil {
				http2.OnError500(w, err, "Can't register job in cron!")
				return true
			}
		}
	}
	return false
}

func (this *CronHandler) Delete(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		http2.OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	if err := this.disable(id); err != nil {
		log.Warnf("Can't unregister job id=%d from cron: %v", id, err)
	}

	if err := this.executionDao.RemoveByJob(id); err != nil {
		http2.OnError500(w, err, fmt.Sprintf("Can't delete executions for job with id %d", id))
		return
	}

	if err := this.jobDao.Remove(id); err != nil {
		http2.OnError500(w, err, fmt.Sprintf("Can't delete job with id %d", id))
		return
	}
}

func (this *CronHandler) Logs(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		http2.OnError(w, err, "Can not parse job id", http.StatusBadRequest)
		return
	}

	execution, found, err := this.executionDao.Get(id)
	if err != nil {
		http2.OnError500(w, err, "Can not get execution with id "+pathParams["id"])
		return
	}
	if !found {
		http2.OnError(w, err, "Execution with id "+pathParams["id"]+" does not exist", http.StatusNotFound)
		return
	}

	r, err := this.logStorage.Read(LogFileKey(execution, pathParams["qualifier"]))
	if err != nil {
		if os.IsNotExist(err) {
			http2.OnError(w, err, "Can not find logs for execution with id "+pathParams["id"]+" qualifier "+pathParams["qualifier"], http.StatusNotFound)
		} else {
			http2.OnError500(w, err, "Can not read logs for execution with id "+pathParams["id"]+" qualifier "+pathParams["qualifier"])
		}
		return
	}
	defer r.Close()

	w.Header().Add("Content-Type", "application/octet-stream")
	io.CopyN(w, r, MAX_LOG_SIZE)
}

func (this *CronHandler) Timeline(w http.ResponseWriter, req *http.Request) {
	fromTimeOffsetStr := req.FormValue("fromTimeOffset")
	fromTimeOffset := -DEFAULT_TIMELINE_DURATION
	if fromTimeOffsetStr != "" {
		o, err := strconv.Atoi(fromTimeOffsetStr)
		if err != nil {
			http2.OnError(w, err, "Can't parse 'fromTimeOffset' int: "+fromTimeOffsetStr, http.StatusBadRequest)
			return
		}
		if o > -1 || o < -72 {
			http2.OnError(w, err, "Incorrect 'fromTimeOffset': "+fromTimeOffsetStr+". Should be betwen -72 and -1 hours", http.StatusBadRequest)
			return
		}
		fromTimeOffset = time.Duration(o) * time.Hour
	}

	jobs, err := this.jobDao.List()
	if err != nil {
		http2.OnError500(w, err, "Can't list jobs")
		return
	}
	jobsById := make(map[int64]cronDao.Job)
	for i := 0; i < len(jobs); i++ {
		jobsById[jobs[i].Id] = jobs[i]
	}

	now := time.Now()
	executions, err := this.executionDao.List(now.Add(fromTimeOffset), now)
	if err != nil {
		http2.OnError500(w, err, "Can't list executions")
		return
	}

	data := make([]Timeline, len(executions))

	for i := 0; i < len(executions); i++ {
		job := jobsById[executions[i].JobId]

		tStart := time.Time(executions[i].Start).Unix()

		tEnd := time.Now().Unix()
		if executions[i].End != nil {
			tEnd = time.Time(*(executions[i].End)).Unix()
		}

		data[i] = Timeline{
			Title:       fmt.Sprintf("%d - %s", job.Id, job.Title),
			Status:      executions[i].Status,
			Start:       tStart,
			End:         max(tStart+1, tEnd),
			ExecutionId: executions[i].Id,
			Manual:      executions[i].Manual,
		}
	}
	this.JsonAnswer(w, data)
}

func (this *CronHandler) Run(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		http2.OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	job, found, err := this.jobDao.Get(id)
	if err != nil {
		http2.OnError500(w, err, fmt.Sprintf("Failed to get job with id %d", id))
		return
	}
	if !found {
		http2.OnError(w, nil, fmt.Sprintf("Can't find job with id %d", id), http.StatusNotFound)
	}

	if err := this.run(job); err != nil {
		http2.OnError500(w, err, "Can't run job")
	}
}

func (this *CronHandler) Version(w http.ResponseWriter, req *http.Request) {
	this.JsonAnswer(w, this.version)
}

func (this *CronHandler) Health(w http.ResponseWriter, req *http.Request) {
	if len(this.registry.failedJobs) > 0 {
		http.Error(w, "Some critical jobs failed", http.StatusInternalServerError)
		this.JsonAnswer(w, this.registry.failedJobs)
	} else {
		this.JsonAnswer(w, "ok")
	}
}

func max(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
