package main

import (
	"bytes"
	"fmt"
	"time"
	"encoding/json"
	"math"
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

func (this *Point) UnmarshalJSON(data []byte) error {
	arr := make([]float64, 2)
	err := json.Unmarshal(data, &arr)
	if (err != nil) {
		return err
	}
	this.x = arr[0]
	this.y = arr[1]
	return nil
}

type JSONTime time.Time

func (t JSONTime)MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02"))
	return []byte(stamp), nil
}

type EventPointType string;

const (
	PHOTO EventPointType = "photo"
	VIDEO EventPointType = "video"
	POST EventPointType = "post"
)

var EventPointAvailableTypes []EventPointType = []EventPointType{PHOTO, VIDEO, POST}

func parseEventPointType(s string) (EventPointType, error) {
	for _,t := range EventPointAvailableTypes {
		if s == string(t) {
			return t, nil
		}
	}
	return "", fmt.Errorf("Unsupported point type %s", s)
}

type EventPoint struct {
	Id      int64 `json:"id"`
	Type    EventPointType `json:"type"`
	Point   Point `json:"point"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Time    JSONTime `json:"time"`
}

type Track struct {
	Id     int64 `json:"id"`
	Title  string `json:"title"`
	Path   []Point `json:"path"`
	Points []EventPoint `json:"points"` // points with articles
}

func (this Track) Bounds(withEventPoints bool) Bbox {
	if len(this.Path) == 0 {
		return Bbox{-180, -90, 180, 90}
	}
	var xMin float64 = 180
	var yMin float64 = 90
	var xMax float64 = -180
	var yMax float64 = -90

	for _, p := range this.Path {
		xMin = math.Min(xMin, p.x)
		yMin = math.Min(yMin, p.y)
		xMax = math.Max(xMax, p.x)
		yMax = math.Max(yMax, p.y)
	}

	if withEventPoints {
		for _,ep := range this.Points {
			xMin = math.Min(xMin, ep.Point.x)
			yMin = math.Min(yMin, ep.Point.y)
			xMax = math.Max(xMax, ep.Point.x)
			yMax = math.Max(yMax, ep.Point.y)
		}
	}

	return Bbox{
		X1:xMin,
		Y1:yMin,
		X2:xMax,
		Y2:yMax,
	}
}

type ExtDataTrack struct {
	Title   string `json:"title"`
	FileIds []string `json:"fileIds"`
}
