package main

import (
	gpx "github.com/ptrv/go-gpx"
	log "github.com/Sirupsen/logrus"
	"io"
)

func parseGpx(reader *io.Reader) ([]Track, error) {
	tracks := make([]Track, 0)

	gpx_data, err := gpx.Parse(*reader)
	if err != nil {
		return nil, err
	}

	log.Infof("%d tracks detected", len(gpx_data.Tracks))
	for _, trk := range gpx_data.Tracks {
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

	log.Infof("%d routes detected", len(gpx_data.Routes))
	for _, route := range gpx_data.Routes {
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
			x:wpt.Lat,
			y:wpt.Lon,
		}
		points = append(points, p)
	}
	return points
}
