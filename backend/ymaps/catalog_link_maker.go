package ymaps

import (
	"github.com/and-hom/wwmap/lib/dao"
	"fmt"
)

type LinkMaker interface {
	Make(spot dao.Spot, river dao.RiverWithSpots) string
}

type NoneLinkMaker struct {

}

func (this NoneLinkMaker)Make(spot dao.Spot, river dao.RiverWithSpots) string {
	return ""
}

type FromSpotLinkMaker struct {

}

func (this FromSpotLinkMaker)Make(spot dao.Spot, river dao.RiverWithSpots) string {
	return spot.Link
}

type WwmapLinkMaker struct {

}

func (this WwmapLinkMaker)Make(spot dao.Spot, river dao.RiverWithSpots) string {
	regionId := river.RegionId
	return fmt.Sprintf("https://wwmap.ru/editor.htm#%d,%d,%d,%d", river.CountryId, regionId, river.Id, spot.Id)
}

type HuskytmLinkMaker struct {

}

func (this HuskytmLinkMaker)Make(spot dao.Spot, river dao.RiverWithSpots) string {
	link, ok := spot.Props[dao.PAGE_LINK_PROP_PREFIX + "huskytm"]
	if ok {
		return link.(string)
	}
	return ""
}