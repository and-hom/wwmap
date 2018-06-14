package util

import "time"

func ZeroDateUTC() time.Time {
	return time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
}