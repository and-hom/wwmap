package command

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/cron2/dao"
	"github.com/and-hom/wwmap/lib/blob"
	"github.com/and-hom/wwmap/lib/util"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const DEFAULT_TTL_DAYS = 120
const LOG_DATE_FORMAT = "2006-01-02 15:04:05-0700"

type CleanerCommand struct {
	ExecutionDao dao.ExecutionDao
	JobDao       dao.JobDao
	LogStorage   blob.BlobStorage
}

func (this CleanerCommand) Create(args string) CommandExecution {
	r, w := io.Pipe()
	ttlDays := make(map[int64]int)
	jobs, err := this.JobDao.List()
	if err != nil {
		log.Error("Can't list jobs", err)
	}
	for _, job := range jobs {
		ttlDays[job.Id] = job.LogsTtlDays
	}

	return cleanerCommandExecution{
		ExecutionDao: this.ExecutionDao,
		LogStorage:   this.LogStorage,
		ErrReader:    r,
		ErrWriter:    w,
		ttlDays:      ttlDays,
	}
}

func (this CleanerCommand) Name() string {
	return "Cleaner"
}

func (this CleanerCommand) String() string {
	return "Cleaner"
}

type cleanerCommandExecution struct {
	ExecutionDao dao.ExecutionDao
	LogStorage   blob.BlobStorage
	ErrReader    io.ReadCloser
	ErrWriter    io.WriteCloser
	ttlDays      map[int64]int
}

func (this cleanerCommandExecution) GetStreamsOrNils() (io.ReadCloser, io.ReadCloser) {
	return nil, this.ErrReader
}

func (this cleanerCommandExecution) Execute() error {
	defer util.DeferCloser(this.ErrWriter)

	maxExecutionIdsByJobId := make(map[int64]int64)
	exCnt := int64(0)
	for jobId, ttlDays := range this.ttlDays {

		deleteBeforeDate := time.Now().Add(time.Duration(-ttlDays) * 24 * time.Hour)
		this.logErr("Remove old executions for job %d before %s", jobId, deleteBeforeDate.Format(LOG_DATE_FORMAT))

		maxId, exCntByJob, err := this.ExecutionDao.RemoveOld(jobId, deleteBeforeDate)
		if err != nil {
			this.logErr("Can't remove old executions: %v", err)
			return err
		}

		maxExecutionIdsByJobId[jobId] = maxId
		exCnt += exCntByJob
		this.logErr("Remove old executions for job %d before id=%d", jobId, maxId)
	}

	keys, err := this.LogStorage.ListIds()
	if err != nil {
		this.logErr("Can't list file storage keys", err)
		return err
	}

	fileCnt := 0
	for _, key := range keys {
		path := strings.Split(key, string(os.PathSeparator))
		if len(path) < 2 {
			this.logErr("Strange log storage key %s", key)
			continue
		}
		jobId, err := strconv.ParseInt(path[0], 10, 64)
		if err != nil {
			this.logErr("Can't detect job id for key ", key, err)
			continue
		}
		executionId, err := strconv.ParseInt(path[1], 10, 64)
		if err != nil {
			this.logErr("Can't detect execution id for key ", key, err)
			continue
		}

		maxId, found := maxExecutionIdsByJobId[jobId]
		if !found {
			maxId = DEFAULT_TTL_DAYS
		}
		if executionId <= maxId {
			this.logErr("Delete log files for key %s", key)
			if err := this.LogStorage.Remove(key); err != nil {
				this.logErr("Can't remove log for key ", key, err)
				continue
			}
			fileCnt++
		}
	}
	this.logErr("Removed %d executions, %d log files", exCnt, fileCnt)

	return nil
}

func (this cleanerCommandExecution) logErr(s string, o ...interface{}) {
	timeStr := time.Now().Format(LOG_DATE_FORMAT + ": ")
	_, err := io.WriteString(this.ErrWriter, timeStr+fmt.Sprintf(s, o...)+"\n")
	if err != nil {
		log.Error(err)
	}
}
