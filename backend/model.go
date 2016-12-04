package main

import (
	"bytes"
	"fmt"
	"time"
)

type Point struct {
	x float64
	y float64
}

func (this Point) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[")
	buffer.WriteString(fmt.Sprint(this.x))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprint(this.y))
	buffer.WriteString("]")
	return buffer.Bytes(), nil
}

type JSONTime time.Time

func (t JSONTime)MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02"))
	return []byte(stamp), nil
}

type EventPoint struct {
	Point Point `json:"point"`
	Id    int64 `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
	Time  JSONTime `json:"time"`
}

type Track struct {
	Id     int64 `json:"-"`
	Title  string `json:"title"`
	Path   []Point `json:"path"`
	Points []EventPoint `json:"points"` // points with articles
}

type ExtDataTrack struct {
	Title   string `json:"title"`
	FileIds []string `json:"fileIds"`
}
