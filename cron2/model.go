package main

import (
	"github.com/and-hom/wwmap/lib/dao"
)

type Job struct {
	dao.IdTitle
	Expr    string `json:"expr"`
	Enabled bool   `json:"enabled"`
	Command string `json:"command"`
}

type Status string

const (
	NEW     Status = "NEW"
	RUNNING Status = "RUNNING"
	DONE    Status = "DONE"
	FAIL    Status = "FAIL"
)

type Execution struct {
	Id     int64         `json:"id"`
	JobId  int64         `json:"job_id"`
	Start  dao.JSONTime  `json:"start"`
	End    *dao.JSONTime `json:"end"`
	Status Status        `json:"status"`
}
