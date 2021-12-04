package tile_data_fetcher

import (
	"github.com/and-hom/wwmap/backend/handler/params"
	"github.com/and-hom/wwmap/backend/handler/toggles"
	"github.com/and-hom/wwmap/lib/geo"
	"net/http"
)

const PREVIEWS_COUNT int = 20

type TileDataFetcher interface {
	Fetch(
		req *http.Request,
		bbox geo.Bbox,
		zoom int,
		requestParams params.Params,
		featureToggles toggles.Toggles,
	) ([]geo.Feature, *http.Request, error)
}

