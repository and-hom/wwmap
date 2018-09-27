package handler

import (
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/and-hom/wwmap/backend/referer"
	"fmt"
)

type App struct {
	Handler
	Storage         Storage
	RiverDao        RiverDao
	WhiteWaterDao   WhiteWaterDao
	ReportDao       ReportDao
	VoyageReportDao VoyageReportDao
	ImgDao          ImgDao
	UserDao         UserDao
	CountryDao      CountryDao
	RegionDao       RegionDao
	AuthProviders	map[AuthProvider]passport.Passport
	RefererStorage  referer.RefererStorage
	ImgUrlBase        string
	ImgUrlPreviewBase string
}

func (this *App) processForWeb(img *Img) {
	if img.Source == IMG_SOURCE_WWMAP {
		img.Url = fmt.Sprintf(this.ImgUrlBase, img.Id)
		img.PreviewUrl = fmt.Sprintf(this.ImgUrlPreviewBase, img.Id)
	}
}


