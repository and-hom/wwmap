package tile_data_fetcher

import (
	"github.com/and-hom/wwmap/backend/handler/params"
	"github.com/and-hom/wwmap/backend/handler/toggles"
	"github.com/and-hom/wwmap/backend/ymaps"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func River(
	tileDao dao.TileDao,
	campDao dao.CampDao,
	linkMaker ymaps.LinkMaker,
	imageProcessor func(img *dao.Img),
	resourceBase string,
) TileDataFetcher {
	return &riverDataFetcher{
		tileDao,
		campDao,
		linkMaker,
		imageProcessor,
		resourceBase,
	}
}

type riverDataFetcher struct {
	TileDao dao.TileDao
	CampDao dao.CampDao

	LinkMaker    ymaps.LinkMaker
	// modifies image hyperlinks
	ImageProcessor func(img *dao.Img)

	ResourceBase string
}

func (this *riverDataFetcher) Fetch(
	req *http.Request,
	bbox geo.Bbox,
	zoom int,
	requestParams params.Params,
	featureToggles toggles.Toggles,
) ([]geo.Feature, *http.Request, error) {

	ctx := req.Context()
	showCamps, ctx := featureToggles.GetShowCamps(ctx)
	if req.Context() != ctx {
		req = req.WithContext(ctx)
	}

	var features []geo.Feature

	river, found, err := this.TileDao.GetRiverById(requestParams.River, PREVIEWS_COUNT)
	if err != nil {
		return nil, req, InternalServerError(err, "Can not read whitewater points for river id %d", requestParams.River)
	}
	if !found {
		return nil, req, NotFound("spots for river %d not found", requestParams.River)
	}

	features = ymaps.WhiteWaterPointsToYmapsNoCluster([]dao.RiverWithSpots{river},
		this.ResourceBase, requestParams.Skip, this.ImageProcessor, this.LinkMaker)

	if showCamps {
		camps, err := this.CampDao.FindWithinBoundsForRiver(bbox, requestParams.River)
		if err != nil {
			log.Error(err)
		} else {
			features = append(features, ymaps.CampsToYmaps(camps, this.ResourceBase, requestParams.Skip)...)
		}
	}

	return features, req, nil
}
