package ymaps

import (
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	"fmt"
	"math"
	"github.com/and-hom/wwmap/lib/model"
	"math/rand"
	"github.com/and-hom/wwmap/backend/clustering"
)

const MAX_CLUSTERS int64 = 8192
const MAX_CLUSTER_ID int64 = int64(math.MaxInt32)
const CLUSTER_CATEGORY_DEFINITING_POINTS_COUNT int = 3

func mkFeature(point WhiteWaterPointWithRiverTitle, river RiverTitle, withDescription bool,
resourcesBase string, processImgForWeb func(img *Img), linkMaker LinkMaker) Feature {
	var description = ""
	if withDescription {
		description = point.ShortDesc
	}

	imgs := make([]Preview, len(point.Images))
	for i := 0; i < len(point.Images); i++ {
		img := &point.Images[i]
		processImgForWeb(img)
		imgs[i] = Preview{
			PreviewUrl:img.PreviewUrl,
			Url:img.Url,
			Source:img.Source,
			RemoteId:img.RemoteId,
		}
	}
	properties := FeatureProperties{
		HintContent: point.Title,
		Id: point.Id,

		Title: point.Title,
		Link: linkMaker.Make(point.WhiteWaterPoint,river),
		ShortDesc: description,
		RiverTitle: point.RiverTitle,
		Images: imgs,

	}
	if point.Category.Category > 0 {
		properties.Category = &point.Category
	}

	return Feature{
		Id:point.Id,
		Geometry:NewYmapsGeoPoint(point.Point),
		Type: FEATURE,
		Properties:properties,
		Options:FeatureOptions{
			IconLayout: IMAGE,
			IconImageHref: CatImg(resourcesBase, point.Category),
			IconImageSize: []int{32, 32},
			IconImageOffset: []int{-16, -16},

			Id: point.Id,
		},
	}
}

func CatImg(resourcesBase string, cat model.SportCategory) string {
	if cat.Impassable() {
		return resourcesBase + "/img/impassable.png"
	} else {
		return fmt.Sprintf(resourcesBase + "/img/cat%d.png", cat.Category)
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

func mkCluster(Id clustering.ClusterId, points []WhiteWaterPointWithRiverTitle) Feature {
	bounds := ClusterGeom(points)
	iconText := points[0].RiverTitle

	return Feature{
		Id: MAX_CLUSTER_ID - rand.Int63n(MAX_CLUSTERS),
		Type: CLUSTER,
		Geometry:NewYmapsGeoPoint(bounds.Center()),
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

func WhiteWaterPointsToYmaps(clusterMaker clustering.ClusterMaker, rivers []RiverTitle, bbox Bbox, zoom int,
resourcesBase string, skipId int64, processImgForWeb func(img *Img), linkMaker LinkMaker) ([]Feature, error) {
	result := make([]Feature, 0)
	for _, river := range rivers {
		riverClusters, err := clusterMaker.Get(river.Id, zoom, bbox)
		if err != nil {
			return []Feature{}, nil
		}

		for id, obj := range riverClusters {
			switch obj.(type) {
			case WhiteWaterPointWithRiverTitle:
				wwp := obj.(WhiteWaterPointWithRiverTitle)
				if wwp.Id != skipId {
					result = append(result, mkFeature(wwp, river, true, resourcesBase, processImgForWeb, linkMaker))
				}
			case clustering.Cluster:
				result = append(result, mkCluster(id, obj.(clustering.Cluster).Points))
			}
		}
	}

	return result, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

