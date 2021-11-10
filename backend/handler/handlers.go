package handler

import (
	"fmt"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/and-hom/wwmap/backend/referer"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/lib/notification"
	"net/http"
)

type App struct {
	Handler
	Storage            Storage
	RiverDao           RiverDao
	TileDao            TileDao
	WhiteWaterDao      WhiteWaterDao
	NotificationDao    NotificationDao
	VoyageReportDao    VoyageReportDao
	ImgDao             ImgDao
	UserDao            UserDao
	CountryDao         CountryDao
	RegionDao          RegionDao
	ChangesLogDao      ChangesLogDao
	MeteoPointDao      MeteoPointDao
	WaterWayDao        WaterWayDao
	WaterWayRefDao     WaterWayRefDao
	AuthProviders      map[AuthProvider]passport.Passport
	RefererStorage     referer.RefererStorage
	ImgUrlBase         string
	ImgUrlPreviewBase  string
	NotificationHelper notification.NotificationHelper
	CampDao            CampDao
	CampPhotoDao       PhotoDao
	CampRateDao        RateDao
}

func (this *App) processForWeb(img *Img) {
	if img.Source == IMG_SOURCE_WWMAP {
		img.Url = fmt.Sprintf(this.ImgUrlBase, img.Id)
		img.PreviewUrl = fmt.Sprintf(this.ImgUrlPreviewBase, img.Id)
	}
}

func  (this *App) experimentalFeaturesEnabled(req *http.Request) (*http.Request, bool, error) {
	user, requestWithUser, authorized, err := GetUser(req, this.UserDao)
	if err!=nil {
		return req, false, err
	}
	if !authorized {
		return req, false, nil
	}
	return requestWithUser, user.ExperimentalFeaures, nil
}
