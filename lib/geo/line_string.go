package geo

func NewPgLineString(points []Point) Geometry {
	return LineString{
		Coordinates: flip(points),
		Type:        LINE_STRING,
	}
}

func NewYmapsLineString(points ...Point) Geometry {
	return LineString{
		Coordinates: points,
		Type:        LINE_STRING,
	}
}

type LineString struct {
	Type        GeometryType `json:"type"`
	Coordinates []Point      `json:"coordinates"`
}

func (this LineString) GetFlippedPath() []Point {
	return flip(this.Coordinates)
}
