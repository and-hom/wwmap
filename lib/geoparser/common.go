package geoparser

import "github.com/and-hom/wwmap/lib/dao"

type GeoParser interface {
	GetTracksAndPoints() ([]dao.Track, []dao.EventPoint, error)
}
