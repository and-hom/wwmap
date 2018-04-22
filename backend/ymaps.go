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
			Type:"Feature",
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
			Type:"Feature",
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
			Type:"Feature",
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
			Type:"Feature",
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

func whiteWaterPointsToYmaps(points []WhiteWaterPoint) []Feature {
	pLength := len(points)

	result := make([]Feature, pLength)
	for i := 0; i < pLength; i++ {
		point := points[i]
		result[i] = Feature{
			Id:point.Id,
			Geometry:NewGeoPoint(point.Point),
			Type:"Feature",
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

	return result
}

func waterwaysToYmaps(waterWays []WaterWay) []Feature {
	features := make([]Feature, len(waterWays))
	for i := 0; i < len(waterWays); i++ {
		points := waterWays[i].Path
		non_zero_points := []Point{}

		for j:=0; j<len(points); j++ {
			if !(points[j].Lon==0 && points[j].Lat==0) {
				non_zero_points = append(non_zero_points, points[j])
			}
		}

		features[i] = Feature{
			Id: waterWays[i].Id,
			Geometry: NewLineString(non_zero_points),
			Type:"Feature",
			Options: FeatureOptions{
				StrokeColor: colorGen(i+1),
			},
		}
	}
	return features
}

func colorGen(i int) string {
	red := (i % 4) * 64
	green := (i % 5) * 51
	blue := (i % 6) * 42
	return fmt.Sprintf("#%02x%02x%02x", red, green, blue)
}

func RoutesToYmaps(route []Route) FeatureCollection {
	var features = []Feature{}
	for i := 0; i < len(route); i++ {
		features = append(features, routeToYmaps(route[i])...)
	}
	return MkFeatureCollection(features)
}

