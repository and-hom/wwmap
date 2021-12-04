package tile_data_fetcher

import (
	"github.com/and-hom/wwmap/backend/clustering"
	"github.com/and-hom/wwmap/backend/handler/params"
	"github.com/and-hom/wwmap/backend/handler/toggles"
	"github.com/and-hom/wwmap/backend/ymaps"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func World(
	tileDao dao.TileDao,
	waterWayDao dao.WaterWayDao,
	campDao dao.CampDao,
	srtmDao dao.SrtmDao,
	clusterMaker clustering.ClusterMaker,
	linkMaker ymaps.LinkMaker,
	imageProcessor func(img *dao.Img),
	resourceBase string,
) TileDataFetcher {
	return &worldDataFetcher{
		tileDao,
		waterWayDao,
		campDao,
		srtmDao,
		clusterMaker,
		linkMaker,
		imageProcessor,
		resourceBase,
	}
}

type worldDataFetcher struct {
	TileDao     dao.TileDao
	WaterWayDao dao.WaterWayDao
	CampDao     dao.CampDao
	SrtmDao     dao.SrtmDao

	ClusterMaker clustering.ClusterMaker
	LinkMaker    ymaps.LinkMaker
	// modifies image hyperlinks
	ImageProcessor func(img *dao.Img)

	ResourceBase string
}

func (this *worldDataFetcher) Fetch(
	req *http.Request,
	bbox geo.Bbox,
	zoom int,
	requestParams params.Params,
	featureToggles toggles.Toggles,
) ([]geo.Feature, *http.Request, error) {
	var features []geo.Feature

	ctx := req.Context()
	showCamps, ctx := featureToggles.GetShowCamps(ctx)
	showUnpublished, ctx := featureToggles.GetShowUnpublished(ctx)
	showSlope, ctx := featureToggles.GetShowSlope(ctx)
	showAltitudeCoverage, ctx := featureToggles.GetAltitudeCoverage(ctx)
	if req.Context() != ctx {
		req = req.WithContext(ctx)
	}

	rivers, err := this.TileDao.ListRiversWithBounds(
		bbox,
		PREVIEWS_COUNT,
		showUnpublished,
	)
	if err != nil {
		return nil, req, InternalServerError(err, "can not read whitewater points for bbox %s", bbox.String())
	}

	riverIds := make([]int64, len(rivers))
	for i := 0; i < len(rivers); i++ {
		riverIds[i] = rivers[i].Id
	}
	//paths, err := this.WaterWayDao.ListByRiverIds(riverIds...)
	//paths, err := this.WaterWayDao.ListByBbox(bbox)
	//if err != nil {
	//	OnError500(w, err, "aaaa")
	//	return
	//}

	paths := []dao.WaterWay{}

	features, err = ymaps.WhiteWaterPointsToYmaps(this.ClusterMaker, rivers, bbox, zoom,
		this.ResourceBase, requestParams.Skip, this.ImageProcessor, this.LinkMaker,
		paths)
	if err != nil {
		return nil, req, InternalServerError(err, "can not cluster: %s %s", bbox.String())
	}

	if showCamps {
		camps, err := this.CampDao.FindWithinBounds(bbox)
		if err != nil {
			log.Error(err)
		} else {
			features = append(features, ymaps.CampsToYmaps(camps, this.ResourceBase, requestParams.Skip)...)
		}
	}
	if showSlope {
		tracks, err := this.WaterWayDao.ListWithHeightsByBbox(bbox)
		if err != nil {
			log.Error(err)
		} else {
			features = append(features, ymaps.TracksToYmaps(tracks, geo.OBJECT_TYPE_SLOPE)...)
		}
	}
	if showAltitudeCoverage {
		coords, err := this.SrtmDao.GetRasterCoords(bbox)
		if err != nil {
			return nil, req, InternalServerError(err, "can't fetch altitude raster coords")
		}

		for _, p := range coords {
			features = append(features, geo.Feature{
				Id:   int64(1000000*p.X + 10000*p.Y),
				Type: geo.FEATURE,
				Geometry: geo.NewYRectangleInt([][]int{
					{p.Y, p.X},
					{p.Y + 1, p.X + 1},
				}),
				Options: geo.FeatureOptions{
					FillColor:   "#ff000055",
					StrokeColor: "#ff0000dd",
				},
				Properties: geo.FeatureProperties{},
			})
		}
	}

	return features, req, nil
}
