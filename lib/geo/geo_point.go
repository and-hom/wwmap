package geo

func NewPgGeoPoint(point Point) Geometry {
	return GeoPoint{
		// flip coordinates for postGIS
		Coordinates: point.Flip(),
		Type:        POINT,
	}
}

func NewYmapsGeoPoint(point Point) Geometry {
	return GeoPoint{
		Coordinates: point,
		Type:        POINT,
	}
}

type GeoPoint struct {
	Type        GeometryType `json:"type"`
	Coordinates Point        `json:"coordinates"`
}
