package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron2/command"
	"github.com/and-hom/wwmap/lib/blob"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Runner struct {
	Job          Job
	Command      command.Command
	ExecutionDao ExecutionDao
	BlobStorage  blob.BlobStorage
	OnComplete   func()
	Manual       bool
}

const (
	STD_OUT = "out"
	STD_ERR = "err"
)

func (this Runner) Run() {
	if this.OnComplete != nil {
		defer this.OnComplete()
	}

	execution, err := this.ExecutionDao.insert(this.Job.Id, this.Manual)
	if err != nil {
		log.Error("Can't insert execution: ", err)
		execution = Execution{Id: -1} // fake execution
	}

	log.Infof("Run job %d execution %d: %s %s %s \"%s\"",
		this.Job.Id, execution.Id, this.Job.Title, this.Job.Expr, this.Job.Command, this.Job.Args)

	cmd := this.Command.Create(this.Job.Args)
	stdout, stderr := cmd.GetStreamsOrNils()

	go this.copyStream(execution, STD_OUT, stdout)()
	go this.copyStream(execution, STD_ERR, stderr)()

	exitStatus := DONE

	if err := cmd.Execute(); err != nil {
		log.Errorf("Execution %d exited: %v", execution.Id, err)
		exitStatus = FAIL
	}

	this.updateStatus(execution, exitStatus)

	if exitStatus == DONE {
		log.Infof("Job %d (execution %d) was successfully ended", this.Job.Id, execution.Id)
	}
}

func (this Runner) updateStatus(e Execution, s Status) {
	if e.Id < 0 {
		log.Error("Try to change status of fake execution - nothing to do")
		return
	}
	if err := this.ExecutionDao.setStatus(e.Id, s); err != nil {
		log.Errorf("Can't set status for execution %d: %v", e.Id, err)
	}
}

func (this Runner) copyStream(execution Execution, qualifier string, stream io.ReadCloser) func() {
	return func() {
		if stream == nil {
			return
		}
		id := LogFileKey(execution, qualifier)
		if err := this.BlobStorage.Store(id, stream); err != nil {
			if strings.HasSuffix(err.Error(), os.ErrClosed.Error()) {
				err = this.BlobStorage.Remove(id)
				if err != nil {
					log.Debug("Can't delete broken log: ", err)
				}
			} else {
				log.Errorf("Can't write %s logs for %d: %v", qualifier, execution.Id, err)
			}
		}
		defer stream.Close()
	}
}

func LogFileKey(execution Execution, qualifier string) string {
	return filepath.Join(fmt.Sprint(execution.JobId), fmt.Sprint(execution.Id), qualifier)
}
