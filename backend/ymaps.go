package main

import (
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	"fmt"
	"math"
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

func mkFeature(point WhiteWaterPointWithRiverTitle, withDescription bool) Feature {
	var description = ""
	if withDescription {
		description = point.ShortDesc
	}

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
			ShortDesc: description,
			RiverName: point.RiverTitle,
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

func ClusterGeom(points []WhiteWaterPointWithRiverTitle) Bbox {
	var minLat = float64(360)
	var minLon = float64(360)
	var maxLat = -float64(360)
	var maxLon = -float64(360)

	for i := 0; i < len(points); i++ {
		lat := points[i].Point.Lat
		lon := points[i].Point.Lon

		minLat = math.Min(minLat, lat)
		minLon = math.Min(minLon, lon)
		maxLat = math.Max(maxLat, lat)
		maxLon = math.Max(maxLon, lon)
	}
	return Bbox{
		X1:minLon,
		Y1: minLat,
		X2:maxLon,
		Y2:maxLat,
	}
}

func mkCluster(Id ClusterId, points []WhiteWaterPointWithRiverTitle) Feature {
	bounds := ClusterGeom(points)
	iconText := points[0].RiverTitle

	return Feature{
		Id: int64(Id.Id) + 10 * Id.RiverId,
		Type: CLUSTER,
		Geometry:NewGeoPoint(bounds.Center()),
		Bbox: bounds.WithMargins(0.05),
		Number: len(points),
		Properties:FeatureProperties{
			IconContent: iconText,

			Title: Id.Title,
		}, Options: FeatureOptions{
			Preset: "islands#greenStretchyIcon",
		},
	}
}

func whiteWaterPointsToYmaps(points []WhiteWaterPointWithRiverTitle, width float64, height float64, zoom int) []Feature {
	by_cluster := clusterMaker.clusterizePoints(points, width, height, zoom)

	result := make([]Feature, 0)
	for id, cluster_points := range by_cluster {
		if len(cluster_points) == 1 &&
			// Show fake cluster for zoom<=SinglePointClusteringMaxZoom on single ww point linked having river id
			(zoom > clusterMaker.SinglePointClusteringMaxZoom || cluster_points[0].RiverId <= 0 || !id.Single) {
			result = append(result, mkFeature(cluster_points[0], true))
		} else {
			result = append(result, mkCluster(id, cluster_points))
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

