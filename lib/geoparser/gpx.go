package geoparser

import (
	gpx "github.com/ptrv/go-gpx"
	log "github.com/Sirupsen/logrus"
	"io"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
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

func (this GpxParser) GetTracksAndPoints() ([]Track, []EventPoint, error) {
	tracks := make([]Track, 0)
	points := make([]EventPoint, 0)

	log.Infof("%d tracks detected", len(this.gpx_data.Tracks))
	for _, trk := range this.gpx_data.Tracks {
		log.Infof("Importing track %s", trk.Name)
		points := make([]Point, 0)
		for _, seg := range trk.Segments {
			points = append(points, convertWaypoints(seg.Waypoints)...)
		}

		startTime, endTime := trk.TimeBounds()
		tracks = append(tracks, Track{
			Title:trk.Name,
			Path:points,
			StartTime:JSONTime(startTime),
			EndTime:JSONTime(endTime),
		})
	}

	for _, wpt := range this.gpx_data.Waypoints {
		log.Infof("Importing waypoint %s", wpt.Name)
		points = append(points, EventPoint{
			Title:wpt.Name,
			Time:JSONTime(wpt.Time()),
			Type:POST,
			Content : wpt.Desc,
			Point:Point{
				Lat:wpt.Lat,
				Lon:wpt.Lon,
			},
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
	return tracks, points, nil
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
