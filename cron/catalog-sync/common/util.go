package common

import (
	"github.com/and-hom/wwmap/lib/util"
	"regexp"
	"time"
)

var zero = util.ZeroDateUTC()

type DateExtractor struct {
	DateFormat string
	DateRegexp *regexp.Regexp
}

func CreateDateExtractor(dateRegexp, dateFormat string) DateExtractor {
	return DateExtractor{
		DateFormat: dateFormat,
		DateRegexp: regexp.MustCompile(dateRegexp),
	}
}

func (this DateExtractor) GetDate(line string) (time.Time, bool) {
	found := this.DateRegexp.FindString(line)
	if found == "" {
		return zero, false
	}
	t, err := time.Parse(this.DateFormat, found)
	if err != nil {
		return zero, false
	}
	return t, true
}

func AppendIfMissing(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}
