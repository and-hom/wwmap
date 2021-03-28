package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/cron2/command"
	"github.com/and-hom/wwmap/cron2/dao"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/robfig/cron"
)

type CronWithRegistry struct {
	cron                     *cron.Cron
	jobRegistry              map[int64]cron.EntryID
	unregisteredReasons      map[int64]string
	manualRunningJobRegistry map[int64]bool
	failedJobs               map[int64]string
	logStorage               blob.BlobStorage
	executionDao             dao.ExecutionDao
	commands                 map[string]command.Command
}

func (this *CronWithRegistry) Register(job dao.Job) error {
	if !job.Enabled {
		log.Infof("Skip diabled job %d %s %s \"%s\"", job.Id, job.Title, job.Expr, job.Command)
		return nil
	}

	c, commandFound := this.commands[job.Command]
	if !commandFound {
		this.unregisteredReasons[job.Id] = "Command not found"
		log.Warnf("Skip job %d with missing command %s", job.Id, job.Command)
		return nil
	}

	log.Infof("Register job %d %s %s \"%s\"", job.Id, job.Title, job.Expr, job.Command)
	runner := Runner{
		Job:          job,
		Command:      c,
		BlobStorage:  this.logStorage,
		ExecutionDao: this.executionDao,
		Manual:       false,
		OnOk: func() {
			delete(this.failedJobs, job.Id)
		},
		OnFail: func(errText string) {
			if job.Critical {
				this.failedJobs[job.Id] = errText
			}
		},
	}
	entryId, err := this.cron.AddFunc(job.Expr, runner.Run)
	if err != nil {
		this.unregisteredReasons[job.Id] = err.Error()
		return err
	}

	this.jobRegistry[job.Id] = entryId
	return nil
}

func (this *CronWithRegistry) Unregister(jobId int64) error {
	log.Infof("Unregister job %d", jobId)
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

func (this *CronWithRegistry) RunNow(job dao.Job) error {
	if _, runnigNow := this.manualRunningJobRegistry[job.Id]; runnigNow {
		return fmt.Errorf("Job %d was started manually and is running just now", job.Id)
	}

	c, commandFound := this.commands[job.Command]
	if !commandFound {
		log.Warnf("Skip job %d with missing command %s", job.Id, job.Command)
		return nil
	}

	log.Infof("Run once just now: job %d %s %s \"%s\"", job.Id, job.Title, job.Expr, job.Command)
	runner := Runner{
		Job:          job,
		Command:      c,
		BlobStorage:  this.logStorage,
		ExecutionDao: this.executionDao,
		OnComplete: func() {
			log.Infof("Completed manually started job %d", job.Id)
			delete(this.manualRunningJobRegistry, job.Id)
		},
		Manual: true,
		OnOk: func() {
			delete(this.failedJobs, job.Id)
		},
	}

	this.manualRunningJobRegistry[job.Id] = true
	go runner.Run()
	return nil
}
