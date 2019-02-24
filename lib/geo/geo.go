package geo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"math"
)

type GeometryType string

type Point struct {
	Lat float64
	Lon float64
}

func (this Point) Flip() Point {
	return Point{
		Lat: this.Lon,
		Lon: this.Lat,
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
	if err != nil {
		return err
	}
	this.Lat = arr[0]
	this.Lon = arr[1]
	return nil
}

func (this Point) DistanceTo(p Point) float64 {
	return math.Sqrt((this.Lat-p.Lat)*(this.Lat-p.Lat) + (this.Lon-p.Lon)*(this.Lon-p.Lon))
}

func (this *Point) String() string {
	return fmt.Sprintf("(lat=%f, lon=%f)", this.Lat, this.Lon)
}

type PointOrLine struct {
	Point *Point   `json:"point,omitempty"`
	Line  *[]Point `json:"line,omitempty"`
}

func (this PointOrLine) Center() Point {
	if this.Point != nil {
		return *this.Point
	}
	if this.Line != nil {
		return (*this.Line)[0] // todo real center ?
	}
	return Point{}
}

func (this PointOrLine) MarshalJSON() ([]byte, error) {
	if this.Point != nil {
		return json.Marshal(this.Point)
	}
	if this.Line != nil {
		return json.Marshal(this.Line)
	}
	return []byte{}, errors.New("Can not serialize empty geo object")
}

func (this *PointOrLine) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	t, err := d.Token()
	if err != nil {
		return err
	}
	if t == json.Delim('[') {
		t2, err := d.Token()
		if err != nil {
			return err
		}
		if t2 == json.Delim('[') {
			if err := json.Unmarshal(data, &this.Line); err != nil {
				return err
			}
		} else {
			this.Point = &Point{}
			if err := json.Unmarshal(data, this.Point); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("Expected [ but found %v", t)
}

func (this PointOrLine) ToPg() interface{} {
	if this.Point != nil {
		return NewPgGeoPoint(*this.Point)
	} else if this.Line != nil {
		return NewPgGeoLine(*this.Line...)
	}
	return nil
}

func (this PointOrLine) Flip() PointOrLine {
	var filppedPointPtr *Point
	if this.Point != nil {
		flipped := (*this.Point).Flip()
		filppedPointPtr = &flipped
	}

	var flippedLine *[]Point
	if this.Line != nil {
		fl := make([]Point, len(*this.Line))
		for i := 0; i < len(*this.Line); i++ {
			fl[i] = (*this.Line)[i].Flip()
		}
		flippedLine = &fl
	}

	return PointOrLine{
		Point: filppedPointPtr,
		Line:  flippedLine,
	}
}

const (
	POINT       GeometryType = "Point"
	RECTANGLE   GeometryType = "Rectangle"
	POLYGON     GeometryType = "Polygon"
	LINE_STRING GeometryType = "LineString"
)

type LineString struct {
	Type        GeometryType `json:"type"`
	Coordinates []Point      `json:"coordinates"`
}

func (this LineString) GetPath() []Point {
	return flip(this.Coordinates)
}

func NewLineString(points []Point) Geometry {
	return LineString{
		Coordinates: flip(points),
		Type:        LINE_STRING,
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
	Coordinates []Point      `json:"coordinates"`
}

func NewYRectangle(points []Point) Geometry {
	return LineString{
		Coordinates: points,
		Type:        RECTANGLE,
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
		Coordinates: points,
		Type:        RECTANGLE,
	}
}

type geoPoint struct {
	Type        GeometryType `json:"type"`
	Coordinates Point        `json:"coordinates"`
}

func NewPgGeoPoint(point Point) Geometry {
	return geoPoint{
		// flip coordinates for postGIS
		Coordinates: point.Flip(),
		Type:        POINT,
	}
}

func NewPgGeoLine(points ...Point) Geometry {
	flippedPoints := make([]Point, len(points))
	for i := 0; i < len(points); i++ {
		flippedPoints[i] = points[i].Flip()
	}
	return LineString{
		Coordinates: flippedPoints,
		Type:        LINE_STRING,
	}
}

func NewYmapsGeoPoint(point Point) Geometry {
	return geoPoint{
		Coordinates: point,
		Type:        POINT,
	}
}

func NewYmapsGeoLine(points ...Point) Geometry {
	return LineString{
		Coordinates: points,
		Type:        LINE_STRING,
	}
}
