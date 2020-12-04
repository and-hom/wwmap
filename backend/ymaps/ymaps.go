package ymaps

import (
	"encoding/json"
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
const CAT_ICON_SIZE = 32

func mkFeature(point Spot, river RiverWithSpots, withDescription bool, resourcesBase string,
	processImgForWeb func(img *Img), linkMaker LinkMaker, visible bool, maxCategory *int) Feature {
	var description = ""
	if withDescription {
		description = point.Description
	}

	imgs := make([]Preview, len(point.Images))
	for i := 0; i < len(point.Images); i++ {
		img := &point.Images[i]
		processImgForWeb(img)
		levelStr := ""
		lvlMap := img.Level
		levelB, err := json.Marshal(lvlMap)
		if err != nil {
			levelStr = "{}"
		} else {
			levelStr = string(levelB)
		}
		avgLevel := 0
		cnt := 0
		for s, v := range img.Level {
			if s != "0" && v > 0 {
				avgLevel += int(v)
				cnt += 1
			}
		}

		if cnt != 0 {
			avgLevel /= cnt
		}

		imgs[i] = Preview{
			Id:         img.Id,
			PreviewUrl: img.PreviewUrl,
			Url:        img.Url,
			Source:     img.Source,
			RemoteId:   img.RemoteId,
			LevelStr:   levelStr,
			Level:      lvlMap,
			AvgLevel:   avgLevel,
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

		RiverCategory: maxCategory,
	}
	if point.Category.Category > 0 {
		properties.Category = &point.Category
	}

	img_x := CAT_ICON_SIZE * (point.Category.Category + 1)
	img_y := 0
	if !visible {
		img_y = CAT_ICON_SIZE * 2
	}

	feature := Feature{
		Id:         point.Id,
		Type:       FEATURE,
		Properties: properties,
		Options: FeatureOptions{
			Id: point.Id,

			IconLayout:    IMAGE,
			IconImageHref: resourcesBase + "/img/categories.png",

			IconImageSize:     []int{CAT_ICON_SIZE, CAT_ICON_SIZE},
			IconImageOffset:   []int{-16, -16},
			IconImageClipRect: [][]int{{img_x, img_y}, {img_x + CAT_ICON_SIZE, img_y + CAT_ICON_SIZE}},
		},
	}

	if point.Point.Line != nil {
		feature.Geometry = NewYmapsLineString(*point.Point.Line...)
		feature.Options.Overlay = "BiPlacemrakOverlay"
		feature.Options.StrokeColor = CatColorWithTransparency(point.Category.Category, visible)
		feature.Options.BorderColor = tubeBorderColor(visible)
	} else if point.Point.Point != nil {
		feature.Geometry = NewYmapsGeoPoint(*point.Point.Point)
	} else {
		logrus.Errorf("Geometry for object id=%d missing", point.Id)
	}

	return feature
}

func tubeBorderColor(visible bool) string {
	if visible {
		return "#444444FF"
	} else {
		return "#44444455"
	}
}

func CatColorWithTransparency(cat int, visible bool) string {
	if visible {
		return catColor(cat) + "CC"
	} else {
		return catColor(cat) + "55"
	}
}

func catColor(cat int) string {
	switch cat {
	case 1:
		return "#00FFF9"
	case 2:
		return "#3CFF00"
	case 3:
		return "#FCFF17"
	case 4:
		return "#FFB100"
	case 5:
		return "#FF0000"
	case 6:
		return "#CC0000"
	}
	return "#BBBBBB"
}

func ClusterGeom(points []Spot) Bbox {
	result := Bbox{
		X1: 360.0,
		Y1: 360.0,
		X2: -360.0,
		Y2: -360.0,
	}

	for i := 0; i < len(points); i++ {
		result.AddPointOrLine(points[i].Point)
	}

	return result
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

func mkCluster(Id clustering.ClusterId, points []Spot, riverTitle string, visible bool) Feature {
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
			Id:       Id.RiverId,
			Color:    CatColorWithTransparency(riverCats.Avg, visible),
		}, Options: options(visible, riverCats),
	}
}

func options(visible bool, riverCats RiverCategoryMetrics) FeatureOptions {
	if visible {
		return FeatureOptions{
			Preset:      categoryClusterIcon(riverCats.Avg),
			StrokeColor: CatColorWithTransparency(riverCats.Avg, visible),
			FillColor:   "white",
		}
	}
	return FeatureOptions{
		Preset:      "wwmap#test", //categoryClusterIcon(riverCats.Avg),
		StrokeColor: "blue",       //CatColor(riverCats.Avg),
		FillColor:   "white",
	}
}

func WhiteWaterPointsToYmaps(clusterMaker clustering.ClusterMaker, rivers []RiverWithSpots, bbox Bbox, zoom int,
	resourcesBase string, skipId int64, processImgForWeb func(img *Img), linkMaker LinkMaker, waterWays []WaterWay) ([]Feature, error) {
	result := make([]Feature, 0)
	for _, river := range rivers {
		riverClusters, err := clusterMaker.Get(river, zoom, bbox)
		if err != nil {
			return []Feature{}, nil
		}
		maxCategory := river.ComputeMaxCategory()

		for id, obj := range riverClusters {
			switch obj.(type) {
			case Spot:
				spot := obj.(Spot)
				if spot.Id != skipId {
					result = append(result, mkFeature(spot, river, true, resourcesBase, processImgForWeb, linkMaker, river.Visible, &maxCategory))
				}
			case clustering.Cluster:
				result = append(result, mkCluster(id, obj.(clustering.Cluster).Points, river.Title, river.Visible))
			}
		}

		for _, waterWay := range waterWays {
			result = append(result, Feature{
				Id:       waterWay.Id,
				Type:     FEATURE,
				Geometry: NewYmapsLineString(waterWay.Path...),
				Properties: FeatureProperties{
					HintContent: fmt.Sprintf("%s/%d", waterWay.Title, waterWay.OsmId),
					Id:          waterWay.Id,

					Title: waterWay.Title,
				},
				Options: FeatureOptions{
					Id: waterWay.Id,
				},
			})
		}
	}

	return result, nil
}

func WhiteWaterPointsToYmapsNoCluster(rivers []RiverWithSpots,
	resourcesBase string, skipId int64, processImgForWeb func(img *Img), linkMaker LinkMaker) []Feature {
	result := make([]Feature, 0)
	for _, river := range rivers {
		for _, spot := range river.Spots {
			if spot.Id != skipId {
				result = append(result, mkFeature(spot, river, true, resourcesBase, processImgForWeb, linkMaker, river.Visible, nil))
			}
		}
	}
	return result
}

func SingleWhiteWaterPointToYmaps(spot Spot, river RiverWithSpots, resourcesBase string, processImgForWeb func(img *Img), linkMaker LinkMaker) ([]Feature, error) {
	return []Feature{mkFeature(spot, river, true, resourcesBase, processImgForWeb, linkMaker, river.Visible, nil)}, nil
}

func CampsToYmaps(camps []Camp, resourcesBase string, skip int64) []Feature {
	result := make([]Feature, 0, len(camps))
	for i := 0; i < len(camps); i++ {
		camp := camps[i]
		if (camp.Id == skip) {
			continue
		}
		result = append(result, Feature{
			Id:       camp.Id,
			Type:     FEATURE,
			Geometry: NewYmapsGeoPoint(camp.Point),
			Options: FeatureOptions{
				Id:              camp.Id,
				IconLayout:      IMAGE,
				IconImageHref:   resourcesBase + "/img/camp.svg",
				IconImageSize:   []int{32, 32},
				IconImageOffset: []int{-16, -16},
			},
			Properties: FeatureProperties{
				Id:          camp.Id,
				Title:       camp.Title,
				HintContent: camp.Title,
			},
		})
	}
	return result
}
