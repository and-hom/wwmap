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

type EventPoint struct {
	Point
	Id int64
	Title string
	Text string
	Time time.Time
}

type Track struct {
	Id    int64 `json:"-"`
	Title string `json:"title"`
	Path  []Point `json:"path"`
	Points  []EventPoint `json:"points"` // points with articles
}

type ExtDataTrack struct {
	Title   string `json:"title"`
	FileIds []string `json:"fileIds"`
}
