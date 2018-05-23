package main

import (
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	"fmt"
	"math"
	"github.com/and-hom/wwmap/lib/model"
	"math/rand"
)

const MAX_CLUSTERS int64 = 8192
const MAX_CLUSTER_ID int64 = int64(math.MaxInt32)
const CLUSTER_CATEGORY_DEFINITING_POINTS_COUNT int = 3

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

func mkFeature(point WhiteWaterPointWithRiverTitle, withDescription bool, resourcesBase string) Feature {
	var description = ""
	if withDescription {
		description = point.ShortDesc
	}

	properties := FeatureProperties{
		HintContent: point.Title,
		Id: point.Id,

		Title: point.Title,
		Link: point.Link,
		ShortDesc: description,
		RiverName: point.RiverTitle,
	}
	if point.Category.Category > 0 {
		properties.Category = &point.Category
	}

	imageHref := fmt.Sprintf(resourcesBase + "/img/cat%d.png", point.Category.Category)
	if point.Category.Category == model.IMPASSABLE {
		imageHref = resourcesBase + "/img/impassable.png"
	}
	return Feature{
		Id:point.Id,
		Geometry:NewGeoPoint(point.Point),
		Type: FEATURE,
		Properties:properties,
		Options:FeatureOptions{
			IconLayout: IMAGE,
			IconImageHref: imageHref,
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

func calculateClusterCategory(points []WhiteWaterPointWithRiverTitle) int {
	cntByCat := make(map[int]int)
	categorizedPointsCount := 0
	for i := 0; i < len(points); i++ {
		currentCat := points[i].Category.Category
		cntByCat[currentCat] += 1
		if currentCat > 0 {
			categorizedPointsCount += 1
		}
	}

	wwCnt := 0
	riverCategory := 0
	definitingPointsCount := min(CLUSTER_CATEGORY_DEFINITING_POINTS_COUNT, categorizedPointsCount)
	for i := 6; i > 0 && wwCnt < definitingPointsCount; i-- {
		wwCnt += cntByCat[i]
		riverCategory = i
	}
	return riverCategory
}

func categoryClusterIcon(category int) string {
	switch category {
	case 6:
		return "islands#redStretchyIcon"
	case 5:
		return "islands#redStretchyIcon"
	case 4:
		return "islands#orangeStretchyIcon"
	case 3:
		return "islands#yellowStretchyIcon"
	case 2:
		return "islands#greenStretchyIcon"
	case 1:
		return "islands#grayStretchyIcon"
	case 0:
		return "islands#lightBlueStretchyIcon"
	default:
		return "islands#lightBlueStretchyIcon"
	}
}

func mkCluster(Id ClusterId, points []WhiteWaterPointWithRiverTitle) Feature {
	bounds := ClusterGeom(points)
	iconText := points[0].RiverTitle


	return Feature{
		Id: MAX_CLUSTER_ID - rand.Int63n(MAX_CLUSTERS),
		Type: CLUSTER,
		Geometry:NewGeoPoint(bounds.Center()),
		Bbox: bounds.WithMargins(0.05),
		Number: len(points),
		Properties:FeatureProperties{
			IconContent: iconText,

			Title: Id.Title,
		}, Options: FeatureOptions{
			Preset: categoryClusterIcon(calculateClusterCategory(points)),
		},
	}
}

func whiteWaterPointsToYmaps(clusterMaker ClusterMaker, points []WhiteWaterPointWithRiverTitle, width float64, height float64, zoom int, resourcesBase string) []Feature {
	by_cluster := clusterMaker.clusterizePoints(points, width, height, zoom)

	result := make([]Feature, 0)
	for id, cluster_points := range by_cluster {
		if len(cluster_points) == 1 &&
		// Show fake cluster for zoom<=SinglePointClusteringMaxZoom on single ww point linked having river id
			(zoom > clusterMaker.SinglePointClusteringMaxZoom || cluster_points[0].RiverId <= 0 || !id.Single) {
			result = append(result, mkFeature(cluster_points[0], true, resourcesBase))
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

func min(x,y int) int {
	if x < y {
		return x
	}
	return y
}

