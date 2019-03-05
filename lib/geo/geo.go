package geo

type GeometryType string

const (
	POINT       GeometryType = "Point"
	RECTANGLE   GeometryType = "Rectangle"
	POLYGON     GeometryType = "Polygon"
	LINE_STRING GeometryType = "LineString"
)

func flip(points []Point) []Point {
	result := make([]Point, len(points))
	for i := 0; i < len(points); i++ {
		result[i] = points[i].Flip()
	}
	return result
}

type PgPointOrLineString struct {
	Coordinates PointOrLine `json:"coordinates"`
}

type PgPolygon struct {
	Coordinates [][]Point `json:"coordinates"`
}
