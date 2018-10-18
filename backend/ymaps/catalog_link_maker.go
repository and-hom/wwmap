package ymaps

import (
	"github.com/and-hom/wwmap/lib/dao"
	"fmt"
)

type LinkMaker interface {
	Make(spot dao.WhiteWaterPoint, river dao.RiverTitle) string
}

type NoneLinkMaker struct {

}

func (this NoneLinkMaker)Make(spot dao.WhiteWaterPoint, river dao.RiverTitle) string {
	return ""
}

type FromSpotLinkMaker struct {

}

func (this FromSpotLinkMaker)Make(spot dao.WhiteWaterPoint, river dao.RiverTitle) string {
	return spot.Link
}

type WwmapLinkMaker struct {

}

func (this WwmapLinkMaker)Make(spot dao.WhiteWaterPoint, river dao.RiverTitle) string {
	regionId := river.Region.Id
	if river.Region.Fake {
		regionId = 0
	}
	return fmt.Sprintf("https://wwmap.ru/editor#%d#%d#%d#%d", river.Region.CountryId, regionId, river.Id, spot.Id)
}

type HuskytmLinkMaker struct {

}

func (this HuskytmLinkMaker)Make(spot dao.WhiteWaterPoint, river dao.RiverTitle) string {
	return
}