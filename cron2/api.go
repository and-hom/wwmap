package main

import (
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/handler"
	http2 "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

const TIMELINE_DURATION = 24 * time.Hour

type CronHandler struct {
	handler.Handler
	jobDao       JobDao
	executionDao ExecutionDao
	userDao      dao.UserDao
	version      string
}

func (this *CronHandler) Init() {
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
	this.Register("/timeline", handler.HandlerFunctions{Get: this.ForRoles(this.Timeline, dao.ADMIN)})
	this.Register("/logs/{id}", handler.HandlerFunctions{Get: this.ForRoles(this.Logs, dao.ADMIN)})
	this.Register("/version", handler.HandlerFunctions{Get: this.Version})
}

func (this *CronHandler) ForRoles(payload handler.HandlerFunction, roles ...dao.Role) handler.HandlerFunction {
	return handler.ForRoles(payload, this.userDao, roles...)
}

func (this *CronHandler) List(w http.ResponseWriter, req *http.Request) {
	jobs, err := this.jobDao.list()
	if err != nil {
		http2.OnError500(w, err, "Can't list jobs")
		return
	}
	this.JsonAnswer(w, jobs)
}

func (this *CronHandler) Get(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		http2.OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	jobs, found, err := this.jobDao.get(id)
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
	job := Job{}
	body, err := handler.DecodeJsonBody(req, &job)
	if err != nil {
		http2.OnError500(w, err, "Can not parse json from request body: "+body)
		return
	}

	if job.Id > 0 {
		err = this.jobDao.update(job)
	} else {
		_, err = this.jobDao.insert(job)
	}
	if err != nil {
		http2.OnError500(w, err, "Can't insert or update")
	}
	this.JsonAnswer(w, true)
}

func (this *CronHandler) Delete(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		http2.OnError(w, err, "Can not parse id", http.StatusBadRequest)
		return
	}

	err = this.executionDao.removeByJob(id)
	if err != nil {
		http2.OnError500(w, err, fmt.Sprintf("Can't delete executions for job with id %d", id))
	}

	err = this.jobDao.remove(id)
	if err != nil {
		http2.OnError500(w, err, fmt.Sprintf("Can't delete job with id %d", id))
	}
}

func (this *CronHandler) Logs(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	id, err := strconv.ParseInt(pathParams["id"], 10, 64)
	if err != nil {
		http2.OnError(w, err, "Can not parse job id", http.StatusBadRequest)
		return
	}

	stub := fmt.Sprintf("Logs for job %d", id)
	this.JsonAnswer(w, []string{stub})
}

func (this *CronHandler) Timeline(w http.ResponseWriter, req *http.Request) {
	jobs, err := this.jobDao.list()
	if err != nil {
		http2.OnError500(w, err, "Can't list jobs")
		return
	}
	jobsById := make(map[int64]Job)
	for i := 0; i < len(jobs); i++ {
		jobsById[jobs[i].Id] = jobs[i]
	}

	now := time.Now()
	executions, err := this.executionDao.list(now.Add(-TIMELINE_DURATION), now)
	if err != nil {
		http2.OnError500(w, err, "Can't list executions")
		return
	}

	data := make([][]interface{}, len(executions))

	for i := 0; i < len(executions); i++ {
		job := jobsById[executions[i].JobId]

		tStart := time.Time(executions[i].Start).Unix()

		tEnd := time.Now().Unix()
		if executions[i].End != nil {
			tEnd = time.Time(*(executions[i].End)).Unix()
		}

		data[i] = []interface{}{
			fmt.Sprintf("%d - %s", job.Id, job.Title),
			executions[i].Status,
			tStart,
			max(tStart, tEnd),
		}
	}
	this.JsonAnswer(w, data)
}

func (this *CronHandler) Version(w http.ResponseWriter, req *http.Request) {
	this.JsonAnswer(w, this.version)
}

func max(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
