package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/blob"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type Runner struct {
	Job          Job
	ExecutionDao ExecutionDao
	BlobStorage  blob.BlobStorage
}

func (this Runner) Run() {
	execution, err := this.ExecutionDao.insert(this.Job.Id)
	if err != nil {
		log.Error("Can't insert execution: ", err)
		execution = Execution{Id: -1} // fake execution
	}
	jobId := fmt.Sprintf("%d", this.Job.Id)
	logId := fmt.Sprintf("%d", execution.Id)

	log.Infof("Run job %d execution %d: %s %s \"%s\"",
		this.Job.Id, execution.Id, this.Job.Title, this.Job.Expr, this.Job.Command)
	cmd := exec.Command("bash", "-c", this.Job.Command)
	stdout := stdOut(cmd.StdoutPipe())
	stderr := stdOut(cmd.StderrPipe())

	go this.copyStream(jobId, logId, "out", stdout)()
	go this.copyStream(jobId, logId, "err", stderr)()

	exitStatus := DONE

	if err := cmd.Start(); err != nil {
		log.Error(err)
		exitStatus = FAIL
	}

	if err := cmd.Wait(); err != nil {
		log.Error(err)
		exitStatus = FAIL
	}

	waitStatus, ok := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if ok && waitStatus.ExitStatus() > 0 {
		log.Errorf("Execution %d exited with status %d", execution.Id, waitStatus.ExitStatus())
		exitStatus = FAIL
	}
	this.updateStatus(execution, exitStatus)
}

func stdOut(pipe io.ReadCloser, err error) io.ReadCloser {
	if err != nil {
		log.Error("Can't get command output stream", err)
		return nil
	}
	return pipe
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

func (this Runner) copyStream(jobId string, logId string, qualifier string, stream io.ReadCloser) func() {
	return func() {
		id := filepath.Join(jobId, logId, qualifier)
		if err := this.BlobStorage.Store(id, stream); err != nil {
			if strings.HasSuffix(err.Error(), os.ErrClosed.Error()) {
				err = this.BlobStorage.Remove(id)
				if err != nil {
					log.Debug("Can't delete broken log: ", err)
				}
			} else {
				log.Errorf("Can't write %s logs for %s: %v", qualifier, logId, err)
			}
		}
		defer stream.Close()
	}
}
