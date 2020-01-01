package dao

import (
	"fmt"
	"github.com/and-hom/wwmap/lib/util"
	"strconv"
	"time"
)

type JSONDate time.Time

func (t JSONDate) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", t.String())
	return []byte(stamp), nil
}

func (t JSONDate) String() string {
	return util.FormatDate(time.Time(t))
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", t.String())
	return []byte(stamp), nil
}

func (t JSONTime) String() string {
	return time.Time(t).Format("2006-01-02 15:04")
}

type JSONUnixTime time.Time

func (t JSONUnixTime) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t JSONUnixTime) String() string {
	return fmt.Sprintf("%d", time.Time(t).Unix())
}

func (this *JSONUnixTime) UnmarshalJSON(data []byte) error {
	t, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*this = JSONUnixTime(time.Unix(t, 0))
	return nil
}
