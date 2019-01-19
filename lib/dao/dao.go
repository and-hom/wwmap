package dao

import (
	. "github.com/and-hom/wwmap/lib/geo"
	_ "github.com/lib/pq"
	"time"
)


type IdEntity interface {
	Remove(id int64, tx interface{}) error
}

type RiverDao interface {
	HasProperties
	IdEntity
	Find(id int64) (River, error)
	ListRiversWithBounds(bbox Bbox, limit int, showUnpublished bool) ([]RiverTitle, error)
	FindTitles(titles []string) ([]RiverTitle, error)
	ListByCountry(countryId int64) ([]RiverTitle, error)
	ListByCountryFull(countryId int64) ([]River, error)
	ListByRegion(regionId int64) ([]RiverTitle, error)
	ListByRegionFull(regionId int64) ([]River, error)
	ListByFirstLetters(query string, limit int) ([]RiverTitle, error)
	Insert(river River) (int64, error)
	Save(river ...River) error
	SetVisible(id int64, visible bool) (error)
	FindByTitlePart(tPart string, limit, offset int) ([]RiverTitle, error)
}

type WhiteWaterDao interface {
	HasProperties
	IdEntity
	InsertWhiteWaterPoints(whiteWaterPoints ...WhiteWaterPoint) error
	InsertWhiteWaterPointFull(whiteWaterPoints WhiteWaterPointFull) (int64, error)
	UpdateWhiteWaterPoints(whiteWaterPoints ...WhiteWaterPoint) error
	UpdateWhiteWaterPointsFull(whiteWaterPoints ...WhiteWaterPointFull) error
	Find(id int64) (WhiteWaterPointWithRiverTitle, error)
	FindFull(id int64) (WhiteWaterPointFull, error)
	ListByBbox(bbox Bbox) ([]WhiteWaterPointWithRiverTitle, error)
	ListByRiver(riverId int64) ([]WhiteWaterPointWithRiverTitle, error)
	ListByRiverFull(riverId int64) ([]WhiteWaterPointFull, error)
	ListByRiverAndTitle(riverId int64, title string) ([]WhiteWaterPointWithRiverTitle, error)
	GetGeomCenterByRiver(riverId int64) (Point, error)
	RemoveByRiver(id int64, tx interface{}) error
	AutoOrderingRiverIds() ([]int64, error)
	DistanceFromBeginning(riverId int64, path []Point) (map[int64]int, error)
	UpdateOrderIdx(idx map[int64]int) error
	FindByTitlePart(tPart string, limit, offset int) ([]WhiteWaterPointWithRiverTitle, error)
}

type NotificationDao interface {
	Add(notification ...Notification) error
	ListUnreadRecipients(nowTime time.Time) ([]NotificationRecipientWithClassifier, error)
	ListUnreadByRecipient(rc NotificationRecipientWithClassifier, limit int) ([]Notification, error)
	MarkRead(notifications []int64) error
}

type WaterWayDao interface {
	AddWaterWays(waterways ...WaterWay) error
	UpdateWaterWay(waterway WaterWay) error
	ForEachWaterWay(transformer func(WaterWay) (WaterWay, error), tmpTable string) error
	DetectForRiver(riverId int64) ([]WaterWay, error)
}

type VoyageReportDao interface {
	UpsertVoyageReports(report ...VoyageReport) ([]VoyageReport, error)
	GetLastId(source string) (interface{}, error)
	AssociateWithRiver(voyageReportId, riverId int64) error
	List(riverId int64, limitByGroup int) ([]VoyageReport, error)
	ForEach(source string, callback func(report *VoyageReport) error) error
	RemoveRiverLink(id int64, tx interface{}) error
}

type ImgDao interface {
	IdEntity
	InsertLocal(wwId int64, _type ImageType, source string, urlBase string, previewUrlBase string, datePublished time.Time) (Img, error)
	Upsert(img ...Img) ([]Img, error)
	Find(id int64) (Img, bool, error)
	List(wwId int64, limit int, _type ImageType, enabledOnly bool) ([]Img, error)
	ListAllBySpot(wwId int64) ([]Img, error)
	ListMainByRiver(wwId int64) ([]Img, error)
	ListAllByRiver(wwId int64) ([]Img, error)
	SetEnabled(id int64, enabled bool) error
	SetMain(spotId int64, id int64) error
	DropMainForSpot(spotId int64) error
	GetMainForSpot(spotId int64) (Img, bool, error)
	RemoveBySpot(spotId int64, tx interface{}) error
	RemoveByRiver(spotId int64, tx interface{}) error
}

type TileDao interface {
	ListRiversWithBounds(bbox Bbox, showUnpublished bool, imgLimit int) ([]RiverWithSpots, error)
	GetRiver(riverId int64, imgLimit int) (RiverWithSpotsExt, error)
}

type WwPassportDao interface {
	Upsert(wwPassport ...WWPassport) error
	GetLastId(source string) (interface{}, error)
}

type UserDao interface {
	CreateIfNotExists(User) (int64, Role, string, bool, error)
	GetRole(provider AuthProvider, extId string) (Role, error)
	List() ([]User, error)
	ListByRole(role Role) ([]User, error)
	SetRole(userId int64, role Role) (Role, Role, error)
	GetBySession(sessionId string) (User, error)
}

type HasProperties interface {
	Props() PropertyManager
}

type CountryDao interface {
	HasProperties
	List() ([]Country, error)
}

type RegionDao interface {
	HasProperties
	Get(id int64) (Region, error)
	GetFake(countryId int64) (Region, bool, error)
	CreateFake(countryId int64) (int64, error)
	List(countryId int64) ([]Region, error)
	ListAllWithCountry() ([]RegionWithCountry, error)
}

type RefererDao interface {
	Put(host string, siteRef SiteRef) error
	List(ttl time.Duration) ([]SiteRef, error)
	RemoveOlderThen(ttl time.Duration) error
}

type ChangesLogDao interface {
	Insert(entry ChangesLogEntry) error
	List(objectType string, objectId int64, limit int) ([]ChangesLogEntry, error)
}