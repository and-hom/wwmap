package geo

import (
	"bytes"
	"fmt"
	"encoding/json"
	"math"
)

type GeometryType string

type Point struct {
	Lat float64
	Lon float64
}

func (this Point) Flip() Point {
	return Point{
		Lat:this.Lon,
		Lon:this.Lat,
	}
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

func (this *Point) DistanceTo(p Point) float64 {
	return math.Sqrt((this.Lat - p.Lat) * (this.Lat - p.Lat) + (this.Lon - p.Lon) * (this.Lon - p.Lon))
}

func (this *Point) String() string {
	return fmt.Sprintf("(lat=%f, lon=%f)", this.Lat, this.Lon)
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

func (this LineString) GetPath() []Point {
	return flip(this.Coordinates)
}

func NewLineString(points []Point) Geometry {
	return LineString{
		Coordinates:flip(points),
		Type:LINE_STRING,
	}
}

func flip(points []Point) []Point {
	result := make([]Point, len(points))
	for i := 0; i < len(points); i++ {
		result[i] = points[i].Flip()
	}
	return result
}

type YRectangle struct {
	Type        GeometryType `json:"type"`
	Coordinates []Point `json:"coordinates"`
}

func NewYRectangle(points []Point) Geometry {
	return LineString{
		Coordinates:points,
		Type:RECTANGLE,
	}
}

func NewYRectangleFromBbox(bbox Bbox) Geometry {
	points := make([]Point, 2)
	points[0] = Point{
		Lon: bbox.X1,
		Lat: bbox.Y1,
	}
	points[1] = Point{
		Lon: bbox.X2,
		Lat: bbox.Y2,
	}
	return LineString{
		Coordinates:points,
		Type:RECTANGLE,
	}
}

type geoPoint struct {
	Type        GeometryType `json:"type"`
	Coordinates Point `json:"coordinates"`
}

func NewPgGeoPoint(point Point) Geometry {
	return geoPoint{
		// flip coordinates for postGIS
		Coordinates: point.Flip(),
		Type:POINT,
	}
}

func NewYmapsGeoPoint(point Point) Geometry {
	return geoPoint{
		Coordinates: point,
		Type:POINT,
	}
}

