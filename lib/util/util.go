package util

import (
	"github.com/Sirupsen/logrus"
	"time"
)

const DEFAULT_LOCATION = "Europe/Moscow"

func GetDefaultLocation() *time.Location {
	l, err := time.LoadLocation(DEFAULT_LOCATION)
	if err != nil {
		logrus.Fatal("Can't load location: ", DEFAULT_LOCATION)
	}
	return l
}

func ToDateInDefaultZone(t time.Time) time.Time {
	return t.In(GetDefaultLocation()).Truncate(24 * time.Hour)
}

func DateEquals(t1 time.Time, t2 time.Time) bool {
	return ToDateInDefaultZone(t1).Equal(ToDateInDefaultZone(t2))
}

func ZeroDateUTC() time.Time {
	return time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func FirstOr(arr []string, _default string) string {
	if len(arr) > 0 {
		return arr[0]
	}
	return _default
}

const DATE_FORMAT = "2006-01-02"

func FormatDate(t time.Time) string {
	return t.Format(DATE_FORMAT)
}

func StringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
