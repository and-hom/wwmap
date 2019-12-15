package dao

import (
	"github.com/and-hom/wwmap/lib/util"
	"github.com/lib/pq"
	"time"
)

func nullDateToZero(date pq.NullTime) time.Time {
	if date.Valid {
		return date.Time
	} else {
		return util.ZeroDateUTC()
	}
}

func nullDateToPtr(date pq.NullTime) *time.Time {
	if date.Valid {
		return &date.Time
	} else {
		return nil
	}
}
