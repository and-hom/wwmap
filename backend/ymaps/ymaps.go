package ymaps

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/backend/clustering"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/model"
	"math"
	"math/rand"
)

const MAX_CLUSTERS int64 = 8192
const MAX_CLUSTER_ID int64 = int64(math.MaxInt32)

func mkFeature(point Spot, river RiverWithSpots, withDescription bool, resourcesBase string,
	processImgForWeb func(img *Img), linkMaker LinkMaker) Feature {
	var description = ""
	if withDescription {
		description = point.Description
	}

	imgs := make([]Preview, len(point.Images))
	for i := 0; i < len(point.Images); i++ {
		img := &point.Images[i]
		processImgForWeb(img)
		imgs[i] = Preview{
			PreviewUrl: img.PreviewUrl,
			Url:        img.Url,
			Source:     img.Source,
			RemoteId:   img.RemoteId,
		}
	}
	properties := FeatureProperties{
		HintContent: point.Title,
		Id:          point.Id,

		Title:      point.Title,
		Link:       linkMaker.Make(point, river),
		ShortDesc:  description,
		RiverTitle: river.Title,
		Images:     imgs,
	}
	if point.Category.Category > 0 {
		properties.Category = &point.Category
	}

	feature := Feature{
		Id:         point.Id,
		Type:       FEATURE,
		Properties: properties,
		Options: FeatureOptions{
			Id: point.Id,

			IconLayout:      IMAGE,
			IconImageHref:   CatImg(resourcesBase, point.Category),
			IconImageSize:   []int{32, 32},
			IconImageOffset: []int{-16, -16},
		},
	}

	if point.Point.Line != nil {
		feature.Geometry = NewYmapsLineString(*point.Point.Line...)
		feature.Options.Overlay = "BiPlacemrakOverlay"
		feature.Options.StrokeColor = CatColor(point.Category.Category)
	} else if point.Point.Point != nil {
		feature.Geometry = NewYmapsGeoPoint(*point.Point.Point)
	} else {
		logrus.Errorf("Geometry for object id=%d missing", point.Id)
	}

	return feature
}

func CatImg(resourcesBase string, cat model.SportCategory) string {
	if cat.Impassable() {
		return resourcesBase + "/img/impassable.png"
	} else {
		return fmt.Sprintf(resourcesBase+"/img/cat%d.png", cat.Category)
	}
}

func CatColor(cat int) string {
	switch cat {
	case 1:
		return "#00FFF9CC"
	case 2:
		return "#3CFF00CC"
	case 3:
		return "#FCFF17CC"
	case 4:
		return "#FFB100CC"
	case 5:
		return "#FF0000CC"
	case 6:
		return "#CC0000CC"
	}
	return "#BBBBBBCC"
}

func ClusterGeom(points []Spot) Bbox {
	var minLat = float64(360)
	var minLon = float64(360)
	var maxLat = -float64(360)
	var maxLon = -float64(360)

	for i := 0; i < len(points); i++ {
		lat := points[i].Point.Center().Lat
		lon := points[i].Point.Center().Lon

		minLat = math.Min(minLat, lat)
		minLon = math.Min(minLon, lon)
		maxLat = math.Max(maxLat, lat)
		maxLon = math.Max(maxLon, lon)
	}
	return Bbox{
		X1: minLon,
		Y1: minLat,
		X2: maxLon,
		Y2: maxLat,
	}
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

func mkCluster(Id clustering.ClusterId, points []Spot, riverTitle string) Feature {
	bounds := ClusterGeom(points)

	riverCats := CalculateClusterCategory(points)

	boundsWithMargins := bounds.WithMargins(0.05)
	return Feature{
		Id:       MAX_CLUSTER_ID - rand.Int63n(MAX_CLUSTERS),
		Type:     CLUSTER,
		Geometry: NewYmapsGeoPoint(bounds.Center()),
		Bbox:     &boundsWithMargins,
		Number:   len(points),
		Properties: FeatureProperties{
			IconContent: riverTitle,

			Title:    Id.Title,
			Category: &model.SportCategory{Category: riverCats.Max},
		}, Options: FeatureOptions{
			Preset: categoryClusterIcon(riverCats.Avg),
		},
	}
}

func WhiteWaterPointsToYmaps(clusterMaker clustering.ClusterMaker, rivers []RiverWithSpots, bbox Bbox, zoom int,
	resourcesBase string, skipId int64, processImgForWeb func(img *Img), linkMaker LinkMaker) ([]Feature, error) {
	result := make([]Feature, 0)
	for _, river := range rivers {
		riverClusters, err := clusterMaker.Get(river, zoom, bbox)
		if err != nil {
			return []Feature{}, nil
		}

		for id, obj := range riverClusters {
			switch obj.(type) {
			case Spot:
				spot := obj.(Spot)
				if spot.Id != skipId {
					result = append(result, mkFeature(spot, river, true, resourcesBase, processImgForWeb, linkMaker))
				}
			case clustering.Cluster:
				result = append(result, mkCluster(id, obj.(clustering.Cluster).Points, river.Title))
			}
		}
	}

	return result, nil
}

func SingleWhiteWaterPointToYmaps(spot Spot, river RiverWithSpots, resourcesBase string, processImgForWeb func(img *Img), linkMaker LinkMaker) ([]Feature, error) {
	return []Feature{mkFeature(spot, river, true, resourcesBase, processImgForWeb, linkMaker)}, nil
}
