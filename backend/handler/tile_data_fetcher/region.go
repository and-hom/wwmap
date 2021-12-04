package tile_data_fetcher

import (
	"github.com/and-hom/wwmap/backend/clustering"
	"github.com/and-hom/wwmap/backend/handler/params"
	"github.com/and-hom/wwmap/backend/handler/toggles"
	"github.com/and-hom/wwmap/backend/ymaps"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"net/http"
)

func Region(
	tileDao dao.TileDao,
	clusterMaker clustering.ClusterMaker,
	linkMaker ymaps.LinkMaker,
	imageProcessor func(img *dao.Img),
	resourceBase string,
	) TileDataFetcher {
	return &regionDataFetcher{
		tileDao,
		clusterMaker,
		linkMaker,
		imageProcessor,
		resourceBase,
	}
}

type regionDataFetcher struct {
	TileDao      dao.TileDao
	ClusterMaker clustering.ClusterMaker
	LinkMaker    ymaps.LinkMaker
	// modifies image hyperlinks
	ImageProcessor func(img *dao.Img)

	ResourceBase string
}

func (this *regionDataFetcher) Fetch(
	req *http.Request,
	bbox geo.Bbox,
	zoom int,
	requestParams params.Params,
	featureToggles toggles.Toggles,
) ([]geo.Feature, *http.Request, error) {

	ctx := req.Context()
	showUnpublished, ctx := featureToggles.GetShowUnpublished(ctx)
	if req.Context() != ctx {
		req = req.WithContext(ctx)
	}

	rivers, err := this.TileDao.ListRiversInRegionWithBounds(
		bbox,
		requestParams.Region,
		PREVIEWS_COUNT,
		showUnpublished,
	)
	if err != nil {
		return nil, req, InternalServerError(err, "Can not read whitewater points for region id %d", requestParams.Region)
	}
	var features []geo.Feature

	features, err = ymaps.WhiteWaterPointsToYmaps(this.ClusterMaker, rivers, bbox, zoom,
		this.ResourceBase, requestParams.Skip, this.ImageProcessor, this.LinkMaker,
		[]dao.WaterWay{})
	if err != nil {
		return nil, req, InternalServerError(err, "Can not cluster %s", bbox.String())
	}

	return features, req, nil
}
