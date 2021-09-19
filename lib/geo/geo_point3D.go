package geo

func NewPgGeoPoint3D(point Point3D) Geometry {
	return GeoPoint3D{
		// flip coordinates for postGIS
		Coordinates: point.Flip(),
		Type:        POINT,
	}
}

func NewYmapsGeoPoint3D(point Point3D) Geometry {
	return GeoPoint3D{
		Coordinates: point,
		Type:        POINT,
	}
}

type GeoPoint3D struct {
	Type        GeometryType `json:"type"`
	Coordinates Point3D        `json:"coordinates"`
}
