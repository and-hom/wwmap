package tile_data_fetcher

import (
	"github.com/and-hom/wwmap/backend/handler/params"
	"github.com/and-hom/wwmap/backend/handler/toggles"
	"github.com/and-hom/wwmap/backend/ymaps"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"net/http"
)

func Spot(
	whiteWaterDao dao.WhiteWaterDao,
	riverDao dao.RiverDao,
	linkMaker ymaps.LinkMaker,
	imageProcessor func(img *dao.Img),
	resourceBase string,
) TileDataFetcher {
	return &spotDataFetcher{
		whiteWaterDao,
		riverDao,
		linkMaker,
		imageProcessor,
		resourceBase,
	}
}

type spotDataFetcher struct {
	WhiteWaterDao dao.WhiteWaterDao
	RiverDao      dao.RiverDao

	LinkMaker ymaps.LinkMaker
	// modifies image hyperlinks
	ImageProcessor func(img *dao.Img)

	ResourceBase string
}

func (this *spotDataFetcher) Fetch(
	req *http.Request,
	_ geo.Bbox,
	_ int,
	requestParams params.Params,
	_ toggles.Toggles,
) ([]geo.Feature, *http.Request, error) {

	var features []geo.Feature

	spot, found, err := this.WhiteWaterDao.Find(requestParams.Only)
	if err != nil {
		return nil, req, InternalServerError(err, "Can not select whitewater point for id %d", requestParams.Only)
	}
	if !found {
		// 404
		return nil, req, NotFound("Can not find whitewater point for id %d", requestParams.Only)
	}
	sp := dao.Spot{
		IdTitle:     spot.IdTitle,
		Description: spot.ShortDesc,
		Link:        spot.Link,
		Point:       spot.Point,
		Images:      spot.Images,
		Category:    spot.Category,
	}

	river, err := this.RiverDao.Find(spot.RiverId)
	if err != nil {
		return nil, req, InternalServerError(err, "Can not find river for id %d", spot.RiverId)
	}
	rws := dao.RiverWithSpots{
		IdTitle:   river.IdTitle,
		CountryId: river.Region.CountryId,
		RegionId:  river.Region.Id,
		Spots:     []dao.Spot{},
		Visible:   river.Visible,
	}

	features, err = ymaps.SingleWhiteWaterPointToYmaps(sp, rws, this.ResourceBase, this.ImageProcessor, this.LinkMaker)

	return features, req, nil
}
