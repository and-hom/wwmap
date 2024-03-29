package util

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

const DEFAULT_LOCATION = "Europe/Moscow"
const USER_AGENT_HEADER = "User-Agent"

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

func PtrToTime(t *time.Time) time.Time {
	if t == nil {
		return ZeroDateUTC()
	}
	return *t
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

func Contains(slice []string, el string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == el {
			return true
		}
	}
	return false
}

func CreateHeadCachingReader(r io.Reader, len int) HeadCachingReader {
	return &headCachingReader{
		r,
		len,
		bytes.NewBuffer(make([]byte, 0)),
	}
}

type HeadCachingReader interface {
	io.Reader
	GetCache() *bytes.Buffer
}

type headCachingReader struct {
	r   io.Reader
	len int
	buf *bytes.Buffer
}

func (t *headCachingReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 && err == nil && t.buf.Len() < t.len {
		t.buf.Write(p[:n])
	}
	return n, err
}

func (t *headCachingReader) GetCache() *bytes.Buffer {
	return t.buf
}

func DeferCloser(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		logrus.Error("Can't close: ", err)
	}
}

func TrimToLengthWithTrailingDots(s string, l int) string {
	if len(s) <= l {
		return s
	}

	if l < 1 {
		return ""
	}
	var dots string
	switch l {
	case 1:
		fallthrough
	case 2:
		dots = ""
	case 3:
		dots = ".."
	default:
		dots = "..."
	}

	textLen := Max(1, l-len(dots))
	return s[:textLen] + dots
}

func Max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
