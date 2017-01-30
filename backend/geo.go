package main

type GeometryType string

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


type GeoParser interface {
	getTracks() ([]Track, error)
}
