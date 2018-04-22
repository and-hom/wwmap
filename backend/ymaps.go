package main

import (
	. "github.com/and-hom/wwmap/backend/dao"
	. "github.com/and-hom/wwmap/backend/geo"
	"fmt"
)

func toYmapsPreset(epType EventPointType) string {
	switch epType {
	case PHOTO:
		return "islands#blueVegetationIcon";
	case VIDEO:
		return "islands#blueVideoIcon";
	case POST:
		return "islands#blueBookIcon";
	}
	return "islands#blueDotIcon";
}

func routeToYmaps(route Route) []Feature {
	pointCount := len(route.Points)
	trackCount := len(route.Tracks)
	featureCount := pointCount + trackCount
	result := make([]Feature, featureCount)

	var i = 0;
	for ; i < pointCount; i++ {
		point := route.Points[i]
		result[i] = Feature{
			Id:point.Id,
			Geometry:NewGeoPoint(point.Point),
			Type: FEATURE,
			Properties:FeatureProperties{
				HintContent: point.Title,
				Id: point.Id,
			},
			Options:FeatureOptions{
				Preset: toYmapsPreset(point.Type),
				Id: point.Id,
			},
		}
	}
	for ; i < featureCount; i++ {
		track := route.Tracks[i - pointCount]
		result[i] = Feature{
			Id:track.Id,
			Geometry:NewLineString(track.Path),
			Type: FEATURE,
		}
	}

	return result
}

func tracksToYmaps(tracks []Track) []Feature {
	tLen := len(tracks)

	result := make([]Feature, tLen)
	for i := 0; i < tLen; i++ {
		track := tracks[i]
		result[i] = Feature{
			Id:track.Id,
			Geometry:NewLineString(track.Path),
			Type: FEATURE,
		}
	}
	return result
}

func pointsToYmaps(points []EventPoint) []Feature {
	pLength := len(points)

	result := make([]Feature, pLength)
	for i := 0; i < pLength; i++ {
		point := points[i]
		result[i] = Feature{
			Id:point.Id,
			Geometry:NewGeoPoint(point.Point),
			Type: FEATURE,
			Properties:FeatureProperties{
				HintContent: point.Title,
				Id: point.Id,
			},
			Options:FeatureOptions{
				Preset: toYmapsPreset(point.Type),
				Id: point.Id,
			},
		}
	}

	return result
}

func clusterizePoints(points []WhiteWaterPoint, width float64, height float64) map[int][]WhiteWaterPoint {
	result := make(map[int][]WhiteWaterPoint)
	for i := 0; i < len(points); i++ {
		result[i] = []WhiteWaterPoint{points[i]}
	}
	return result
}

func mkFeature(point WhiteWaterPoint) Feature {
	return Feature{
		Id:point.Id,
		Geometry:NewGeoPoint(point.Point),
		Type: FEATURE,
		Properties:FeatureProperties{
			HintContent: point.Title,
			Id: point.Id,

			Title: point.Title,
			Category: point.Category,
			Link: point.Link,
			ShortDesc: point.ShortDesc,
		},
		Options:FeatureOptions{
			IconLayout: IMAGE,
			IconImageHref: fmt.Sprintf("img/cat%d.png", point.Category.Category),
			IconImageSize: []int{32, 32},
			IconImageOffset: []int{-16, -16},

			Id: point.Id,
		},
	}
}

func whiteWaterPointsToYmaps(points []WhiteWaterPoint, width float64, height float64) []Feature {
	by_cluster := clusterizePoints(points, width, height)

	result := make([]Feature, 0)
	for _, cluste_points := range by_cluster {
		if len(cluste_points) == 1 {
			result = append(result, mkFeature(points[0]))
		}
	}

	return result
}

func RoutesToYmaps(route []Route) FeatureCollection {
	var features = []Feature{}
	for i := 0; i < len(route); i++ {
		features = append(features, routeToYmaps(route[i])...)
	}
	return MkFeatureCollection(features)
}

