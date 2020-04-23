package main

import "github.com/and-hom/wwmap/cron2/dao"

type Timeline struct {
	Title       string     `json:"title"`
	Status      dao.Status `json:"status"`
	Start       int64      `json:"start"`
	End         int64      `json:"end"`
	ExecutionId int64      `json:"execution_id"`
	Manual      bool       `json:"manual"`
}
