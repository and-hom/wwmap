package handler

import (
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/and-hom/wwmap/backend/referer"
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
}


