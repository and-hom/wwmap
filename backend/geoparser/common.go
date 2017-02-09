package geoparser

import "github.com/and-hom/wwmap/backend/dao"

type GeoParser interface {
	GetTracks() ([]dao.Track, error)
}
