package main

import (
	"strings"
	"fmt"
	"strconv"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
)

type Bbox struct {
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
}

func NewBbox(data string) (Bbox, error) {
	parts := strings.Split(data, ",")
	if len(parts) != 4 {
		return Bbox{}, fmt.Errorf("%s is illegal bbox representation. Length sould be equals %d", data, len(parts))
	}
	partsF := make([]float64, 4)
	for i := 0; i < 4; i++ {
		partF, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return Bbox{}, fmt.Errorf("%s is illegal bbox representation. Length sould be equals %d", data, len(parts))
		}
		partsF[i] = partF
	}

	return Bbox{
		X1:partsF[0],
		Y1:partsF[1],
		X2:partsF[2],
		Y2:partsF[3],
	}, nil
}

type GeometryType string

const (
	POINT GeometryType = "Point"
)

type Geometry interface {

}

type lineString struct {
	Type        GeometryType `json:"type"`
	Coordinates []Point `json:"coordinates"`
}

func NewLineString(points []Point) Geometry {
	return lineString{
		Coordinates:points,
		Type:"LineString",
	}
}

type yPoint struct {
	Type        GeometryType `json:"type"`
	Coordinates Point `json:"coordinates"`
}

func NewYPoint(point Point) Geometry {
	return yPoint{
		Coordinates:point,
		Type:"Point",
	}
}

type FeatureProperties struct {
	BalloonContent string `json:"balloonContent,omitempty"`
	ClusterCaption string `json:"clusterCaption,omitempty"`
	HintContent    string `json:"hintContent,omitempty"`
}

type Feature struct {
	Type       string `json:"type"`
	Id         int64 `json:"id"`
	Geometry   Geometry `json:"geometry"`
	Properties FeatureProperties `json:"properties,omitempty"`
}

type FeatureCollection struct {
	Features []Feature `json:"features"`
	Type     string `json:"type"`
}

func trackToYmaps(track Track) []Feature {
	pointCount := len(track.Points)
	result := make([]Feature, pointCount + 1)
	for i := 0; i < pointCount; i++ {
		point := track.Points[i]
		result[i] = Feature{
			Id:point.Id,
			Geometry:NewYPoint(point.Point),
			Type:"Feature",
			Properties:FeatureProperties{
				BalloonContent:	point.Text,
				HintContent: point.Title,
			},
		}
	}
	result[pointCount] = Feature{
		Id:track.Id,
		Geometry:NewLineString(track.Path),
		Type:"Feature",
	}
	return result
}

func TracksToYmaps(tracks []Track) FeatureCollection {
	featuresForTracks := make([][]Feature, len(tracks))
	for i := 0; i < len(tracks); i++ {
		featuresForTracks[i] = trackToYmaps(tracks[i])
	}
	features := flatten(featuresForTracks)
	return FeatureCollection{
		Features:features,
		Type:"FeatureCollection",
	}
}

func flatten(arrs [][]Feature) []Feature {
	totalSize := 0
	for i := 0; i < len(arrs); i++ {
		totalSize += len(arrs[i])
	}
	result := make([]Feature, totalSize)
	pos := 0
	for i := 0; i < len(arrs); i++ {
		for j := 0; j < len(arrs[i]); j++ {
			result[pos] = arrs[i][j]
			pos++
		}
	}
	return result
}

func (f FeatureCollection) Json() string {
	bytes, err := json.Marshal(f)
	if err != nil {
		log.Errorf("Can not serialize object %v: %s", f, err.Error())
		return "{}"
	}
	return string(bytes)
}
