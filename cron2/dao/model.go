package dao

import (
	"github.com/and-hom/wwmap/lib/dao"
)

type Job struct {
	dao.IdTitle
	Expr     string `json:"expr"`
	Enabled  bool   `json:"enabled"`
	Critical bool   `json:"critical"`
	Command  string `json:"command"`
	Args     string `json:"args"`
}

type Status string

const (
	NEW     Status = "NEW"
	RUNNING Status = "RUNNING"
	DONE    Status = "DONE"
	FAIL    Status = "FAIL"
	ORPHAN  Status = "ORPHAN" // If app exited and execution is not under the monitoring
)

type Execution struct {
	Id     int64         `json:"id"`
	JobId  int64         `json:"job_id"`
	Start  dao.JSONTime  `json:"start"`
	End    *dao.JSONTime `json:"end"`
	Status Status        `json:"status"`
	Manual bool          `json:"manual"`
}
