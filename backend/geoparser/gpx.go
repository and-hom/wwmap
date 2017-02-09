package geoparser

import (
	gpx "github.com/ptrv/go-gpx"
	log "github.com/Sirupsen/logrus"
	"io"
	. "github.com/and-hom/wwmap/backend/dao"
	. "github.com/and-hom/wwmap/backend/geo"
)

type GpxParser struct {
	gpx_data *gpx.Gpx
}

func InitGpxParser(reader io.Reader) (*GpxParser, error) {
	gpx_data, err := gpx.Parse(reader)
	if err != nil {
		return nil, err
	}
	parser := GpxParser{
		gpx_data:gpx_data,
	}
	return &parser, nil
}

func (this GpxParser) GetTracks() ([]Track, error) {
	tracks := make([]Track, 0)

	log.Infof("%d tracks detected", len(this.gpx_data.Tracks))
	for _, trk := range this.gpx_data.Tracks {
		log.Infof("Importing %s", trk.Name)
		points := make([]Point, 0)
		for _, seg := range trk.Segments {
			points = append(points, convertWaypoints(seg.Waypoints)...)
		}

		tracks = append(tracks, Track{
			Title:trk.Name,
			Path:points,
		})
	}

	log.Infof("%d routes detected", len(this.gpx_data.Routes))
	for _, route := range this.gpx_data.Routes {
		log.Infof("Importing %s", route.Name)
		tracks = append(tracks, Track{
			Title:route.Name,
			Path:convertWaypoints(route.Waypoints),
		})
	}
	return tracks, nil
}

func convertWaypoints(wpts gpx.Waypoints) []Point {
	points := make([]Point, 0)

	for _, wpt := range wpts {
		p := Point{
			Lat:wpt.Lat,
			Lon:wpt.Lon,
		}
		points = append(points, p)
	}
	return points
}
