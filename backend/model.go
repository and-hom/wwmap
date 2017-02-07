package main

import (
	"bytes"
	"fmt"
	"time"
	"encoding/json"
	"math"
)

type Point struct {
	lat float64
	lon float64
}

func (this Point) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[")
	buffer.WriteString(fmt.Sprint(this.lat))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprint(this.lon))
	buffer.WriteString("]")
	return buffer.Bytes(), nil
}

func (this *Point) UnmarshalJSON(data []byte) error {
	arr := make([]float64, 2)
	err := json.Unmarshal(data, &arr)
	if (err != nil) {
		return err
	}
	this.lat = arr[0]
	this.lon = arr[1]
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
	for _, t := range EventPointAvailableTypes {
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

type TrackType string;

const (
	UNKNOWN TrackType = ""
	PEDESTRIAN TrackType = "pd"
	BIKE TrackType = "bk"
	WATER TrackType = "ww"
)

var TrackAvailableTypes []TrackType = []TrackType{PEDESTRIAN, BIKE, WATER, UNKNOWN}

func parseTrackType(s string) (TrackType, error) {
	for _, t := range TrackAvailableTypes {
		if s == string(t) {
			return t, nil
		}
	}
	return "", fmt.Errorf("Unsupported track type %s", s)
}

type Track struct {
	Id    int64 `json:"id"`
	Title string `json:"title"`
	Path  []Point `json:"path"`
	Type  TrackType `json:"type"`
}

func (this Track) Bounds() Bbox {
	if len(this.Path) == 0 {
		return Bbox{-180, -90, 180, 90}
	}
	var xMin float64 = 180
	var yMin float64 = 90
	var xMax float64 = -180
	var yMax float64 = -90

	for _, p := range this.Path {
		xMin = math.Min(xMin, p.lat)
		yMin = math.Min(yMin, p.lon)
		xMax = math.Max(xMax, p.lat)
		yMax = math.Max(yMax, p.lon)
	}

	return Bbox{
		X1:xMin,
		Y1:yMin,
		X2:xMax,
		Y2:yMax,
	}
}

type Route struct {
	Id     int64 `json:"id"`
	Title  string `json:"title"`
	Tracks []Track `json:"tracks"`
	Points []EventPoint `json:"points"` // points with articles
}

func Bounds(tracks []Track, points []EventPoint) Bbox {
	var xMin float64 = 180
	var yMin float64 = 90
	var xMax float64 = -180
	var yMax float64 = -90

	for _, tr := range tracks {
		trackBounds := tr.Bounds()
		xMin = math.Min(xMin, trackBounds.X1)
		yMin = math.Min(yMin, trackBounds.Y1)
		xMax = math.Max(xMax, trackBounds.X2)
		yMax = math.Max(yMax, trackBounds.Y2)
	}
	for _, ep := range points {
		xMin = math.Min(xMin, ep.Point.lat)
		yMin = math.Min(yMin, ep.Point.lon)
		xMax = math.Max(xMax, ep.Point.lat)
		yMax = math.Max(yMax, ep.Point.lon)
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
