package geoparser

import "github.com/and-hom/wwmap/backend/dao"

type GeoParser interface {
	GetTracksAndPoints() ([]dao.Track, []dao.EventPoint, error)
}
