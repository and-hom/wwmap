package main

import "github.com/and-hom/wwmap/cron2/dao"

type Timeline struct {
	Title         string     `json:"title"`
	Status        dao.Status `json:"status"`
	Start         int64      `json:"start"`
	End           int64      `json:"end"`
	ExecutionId   int64      `json:"execution_id"`
	Manual        bool       `json:"manual"`
	SquashedCount int        `json:"squashed_count"`
	jobId         int64      `json:"-"`
	lastElStart   int64      `json:"-"`
}

type JobDto struct {
	dao.Job
	Registered         bool   `json:"registered"`
	UnregisteredReason string `json:"unregistered_reason,omitempty"`
}
