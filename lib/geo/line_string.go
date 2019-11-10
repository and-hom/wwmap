package geo

import "math"

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

func (this LineString) GetBounds(koeff float64) Bbox {
	if len(this.Coordinates) == 0 {
		return Bbox{}
	}
	min := Point{math.MaxFloat32, math.MaxFloat32}
	max := Point{-math.MaxFloat32, -math.MaxFloat32}
	for i := 0; i < len(this.Coordinates); i++ {
		p := this.Coordinates[i]

		if min.Lat > p.Lat {
			min.Lat = p.Lat
		}
		if min.Lon > p.Lon {
			min.Lon = p.Lon
		}

		if max.Lat < p.Lat {
			max.Lat = p.Lat
		}
		if max.Lon < p.Lon {
			max.Lon = p.Lon
		}
	}

	xc := (min.Lon + max.Lon) / 2
	yc := (min.Lat + max.Lat) / 2

	dx := max.Lon - xc
	dy := max.Lat - yc
	return Bbox{
		xc - koeff*dx, yc - koeff*dy,
		xc + koeff*dx, yc + koeff*dy,
	}
}
