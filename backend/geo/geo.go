package geo

import (
	"bytes"
	"fmt"
	"encoding/json"
)

type GeometryType string

type Point struct {
	Lat float64
	Lon float64
}

func (this Point) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[")
	buffer.WriteString(fmt.Sprint(this.Lat))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprint(this.Lon))
	buffer.WriteString("]")
	return buffer.Bytes(), nil
}

func (this *Point) UnmarshalJSON(data []byte) error {
	arr := make([]float64, 2)
	err := json.Unmarshal(data, &arr)
	if (err != nil) {
		return err
	}
	this.Lat = arr[0]
	this.Lon = arr[1]
	return nil
}

const (
	POINT GeometryType = "Point"
	RECTANGLE GeometryType = "Rectangle"
	POLYGON GeometryType = "Polygon"
	LINE_STRING GeometryType = "LineString"
)

type LineString struct {
	Type        GeometryType `json:"type"`
	Coordinates []Point `json:"coordinates"`
}

func NewLineString(points []Point) Geometry {
	return LineString{
		Coordinates:points,
		Type:LINE_STRING,
	}
}


type geoPoint struct {
	Type        GeometryType `json:"type"`
	Coordinates Point `json:"coordinates"`
}

func NewGeoPoint(point Point) Geometry {
	return geoPoint{
		Coordinates:point,
		Type:POINT,
	}
}

