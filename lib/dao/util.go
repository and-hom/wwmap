package dao

import (
	"github.com/and-hom/wwmap/lib/util"
	"github.com/lib/pq"
	"regexp"
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

func zeroToPqDate(t time.Time) pq.NullTime {
	if t == zeroDate {
		return pq.NullTime{Valid: false}
	} else {
		return pq.NullTime{Valid: true, Time: t}
	}
}

func nullPtrToPqDate(t *time.Time) pq.NullTime {
	if t == nil {
		return pq.NullTime{Valid: false}
	} else {
		return pq.NullTime{Valid: true, Time: *t}
	}
}

var eYoRepl = regexp.MustCompile(`(?i)ё`)
