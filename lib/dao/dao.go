package dao

import (
	. "github.com/and-hom/wwmap/lib/geo"
	_ "github.com/lib/pq"
	"io"
	"time"
)

type IdEntity interface {
	Remove(id int64, tx interface{}) error
}

type RiverDao interface {
	HasProperties
	IdEntity
	Find(id int64) (River, error)
	FindForImage(imgId int64) (River, error)
	FindForSpot(spotId int64) (River, error)
	ListRiversWithBounds(bbox Bbox, limit int, showUnpublished bool) ([]RiverTitle, error)
	FindTitles(titles []string) ([]RiverTitle, error)
	ListAll() ([]RiverTitle, error)
	ListByCountry(countryId int64) ([]RiverTitle, error)
	ListByCountryFull(countryId int64) ([]River, error)
	ListByRegion(regionId int64) ([]RiverTitle, error)
	ListByRegionFull(regionId int64) ([]River, error)
	ListByFirstLetters(query string, limit int) ([]RiverTitle, error)
	Insert(river River) (int64, error)
	SaveFull(river ...River) error
	Save(river ...RiverTitle) error
	SetVisible(id int64, visible bool) error
	FindByTitlePart(tPart string, regionId int64, countryId int64, limit, offset int, showUnpublished bool) ([]RiverTitle, error)
	GetParentIds(riverIds []int64) (map[int64]RiverParentIds, error)
	CountByRegion(regionId int64) (int, error)
}

type WhiteWaterDao interface {
	HasProperties
	IdEntity
	InsertWhiteWaterPoints(whiteWaterPoints ...WhiteWaterPoint) error
	InsertWhiteWaterPointFull(whiteWaterPoints WhiteWaterPointFull, tx interface{}) (int64, error)
	UpdateWhiteWaterPoints(whiteWaterPoints ...WhiteWaterPoint) error
	UpdateWhiteWaterPointsFull(whiteWaterPoints ...WhiteWaterPointFull) error
	UpdateWhiteWaterPointFull(whiteWaterPoints WhiteWaterPointFull, tx interface{}) error
	Find(id int64) (WhiteWaterPointWithRiverTitle, bool, error)
	FindFull(id int64) (WhiteWaterPointFull, error)
	ListByBbox(bbox Bbox) ([]WhiteWaterPointWithRiverTitle, error)
	ListByRiver(riverId int64) ([]WhiteWaterPointWithRiverTitle, error)
	ListByRiverFull(riverId int64) ([]WhiteWaterPointFull, error)
	ListByRiverAndTitle(riverId int64, title string) ([]WhiteWaterPointWithRiverTitle, error)
	GetGeomCenterByRiver(riverId int64) (*Point, error)
	GetRiverBounds(riverId int64) (*Bbox, error)
	RemoveByRiver(id int64, tx interface{}) error
	AutoOrderingRiverIds() ([]int64, error)
	DistanceFromBeginning(riverId int64, path []Point) (map[int64]int, error)
	UpdateOrderIdx(idx map[int64]int) error
	FindByTitlePart(tPart string, regionId int64, countryId int64, limit, offset int, showUnpublished bool) ([]WhiteWaterPointWithRiverTitle, error)
	GetParentIds(spotIds []int64) (map[int64]SpotParentIds, error)
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
	UnlinkRiver(id int64, tx interface{}) error
	BindWaterwaysToRivers() error
	ListByRiverIds(riverIds ...int64) ([]WaterWay, error)
	ListWithRiver() ([]WaterWay4Router, error)
	ListByBbox(bbox Bbox) ([]WaterWay, error)
	ListByBboxNonFilpped(bbox Bbox) ([]WaterWay4Router, error)
	ListByRiverIdNonFlipped(riverId int64) ([]WaterWay4Router, error)
	ListWithHeightsByBbox(bbox Bbox) ([]WaterWayWithHeight, error)
	List(limit int, offset int) ([]WaterWay4PathCorrection, error)
	PathSimplifiedPersister() (PathSimplifiedPersister, error)
	PathHeightPersister() (PathHeightPersister, error)
}

type PathSimplifiedPersister interface {
	io.Closer
	Add(id int64, pathSimplified []Point) error
}

type PathHeightPersister interface {
	io.Closer
	Add(id int64, pathHeight []int32, dists []float64) error
}

type WaterWayOsmRefDao interface {
	Insert(refs ...WaterWayOsmRef) error
}

type WaterWayRefDao interface {
	RefsById() (map[int64][]int64, error)
}

type ImgDao interface {
	IdEntity
	HasProperties
	InsertLocal(wwId int64, _type ImageType, source string, datePublished time.Time, date *time.Time, level map[string]int8) (Img, error)
	Upsert(img ...Img) ([]Img, error)
	Find(id int64) (Img, bool, error)
	List(wwId int64, limit int, _type ImageType, enabledOnly bool) ([]Img, error)
	ListExt(wwId int64, limit int, _type ImageType, enabledOnly bool) ([]ImgExt, error)
	ListAllBySpot(wwId int64) ([]Img, error)
	ListMainByRiver(riverId int64) ([]Img, error)
	ListAllByRiver(riverId int64) ([]Img, error)
	SetEnabled(id int64, enabled bool) error
	SetDateAndLevel(id int64, date time.Time, level map[string]int8) error
	SetManualLevel(id int64, level int8) (map[string]int8, error)
	ResetManualLevel(id int64) (map[string]int8, error)
	SetMain(spotId int64, id int64) error
	DropMainForSpot(spotId int64) error
	GetMainForSpot(spotId int64) (Img, bool, error)
	RemoveBySpot(spotId int64, tx interface{}) error
	RemoveByRiver(spotId int64, tx interface{}) error
	GetParentIds(imgIds []int64) (map[int64]ImageParentIds, error)
}

type TileDao interface {
	ListRiversWithBounds(bbox Bbox, imgLimit int, showUnpublished bool) ([]RiverWithSpots, error)
	ListRiversInRegionWithBounds(bbox Bbox, regionId int64, imgLimit int, showUnpublished bool) ([]RiverWithSpots, error)
	ListRiversInCountryWithBounds(bbox Bbox, countryId int64, imgLimit int, showUnpublished bool) ([]RiverWithSpots, error)
	GetRiverById(riverId int64, imgLimit int) (RiverWithSpots, bool, error)
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
	SetExperimentalFeatures(userId int64, enable bool) (bool, bool, error)
	GetBySession(sessionId string) (User, bool, error)
}

type HasProperties interface {
	Props() PropertyManager
}

type CountryDao interface {
	HasProperties
	List() ([]Country, error)
	Get(id int64) (Country, bool, error)
	GetByCode(code string) (Country, bool, error)
}

type RegionDao interface {
	HasProperties
	Get(id int64) (Region, bool, error)
	GetFake(countryId int64) (Region, bool, error)
	CreateFake(countryId int64) (int64, error)
	List(countryId int64) ([]Region, error)
	ListAllWithCountry() ([]RegionWithCountry, error)
	Save(region ...Region) error
	Insert(region Region) (int64, error)
	Remove(id int64) error
	GetParentIds(regionIds []int64) (map[int64]RegionParentIds, error)
}

type RefererDao interface {
	Put(host string, siteRef SiteRef) error
	List(ttl time.Duration) ([]SiteRef, error)
	RemoveOlderThen(ttl time.Duration) error
}

type ChangesLogDao interface {
	Insert(entry ChangesLogEntry) error
	List(objectType string, objectId int64, limit int) ([]ChangesLogEntry, error)
	ListAll(limit int) ([]ChangesLogEntry, error)
	ListAllWithUserInfo(limit int) ([]ChangesLogEntryWithUseInfo, error)
	ListWithUserInfo(objectType string, objectId int64, limit int) ([]ChangesLogEntryWithUseInfo, error)
	ListAllTimeRange(fromInclude time.Time, toExclude time.Time, limit int) ([]ChangesLogEntry, error)
}

type MeteoDao interface {
	Insert(meteo Meteo) error
}

type MeteoPointDao interface {
	Find(id int64) (MeteoPoint, error)
	Insert(meteo MeteoPoint) (MeteoPoint, error)
	List() ([]MeteoPoint, error)
}

type LevelSensorDao interface {
	Find(id string) (LevelSensor, error)
	List() ([]LevelSensor, error)
	SetGraduation(id string, graduation [LEVEL_GRADUATION]int) error
	CreateIfMissing(id string) error
}

type LevelDao interface {
	Insert(level Level) error
	ListBySensorAndDate(fromDate time.Time, toDate time.Time) (map[string]map[string]Level, error)
	ListForSensor(sensorId string) ([]Level, error)
	GetDailyLevelBetweenDates(sensorId string, from time.Time, to time.Time) ([]Level, error)
	RemoveNullsBefore(fromDate JSONDate) error
}

type RateDao interface {
	RemoveByRefId(refId int64, tx interface{}) error
}

type PhotoDao interface {
	RemoveByRefId(refId int64, tx interface{}) error
}

type DbVersionDao interface {
	GetDbVersion() (int, error)
}

type RiverLinksDao interface {
	SetLinksForRiver(riverId int64, entityIds []int64) error
    GetIdsForRiver(riverId int64) ([]int64, error)
	ExistsByRiver(riverId int64) (bool, error)
}

type TransferDao interface {
	RiverLinksDao
	List(withRivers bool) ([]Transfer, error)
	Find(id int64) (Transfer, bool, error)
	ByRiver(riverId int64) ([]Transfer, error)
	Insert(transfer Transfer) (int64, error)
	Update(transfer Transfer) error
	Remove(id int64) error
}

type CampDao interface {
	RiverLinksDao
	List(withRivers bool) ([]Camp, error)
	ByRiver(riverId int64) ([]Camp, error)
	Insert(camp Camp) (int64, error)
	InsertMultiple(camp ...Camp) ([]int64, error)
	Update(camp Camp) error
	Find(id int64) (Camp, bool, error)
	Remove(id int64, tx interface{}) error
	FindWithinBounds(bbox Bbox) ([]Camp, error)
	FindWithinBoundsForRiver(bbox Bbox, riverId int64) ([]Camp, error)
}

type VoyageReportDao interface {
	RiverLinksDao
	List(withRivers bool, withRemoved bool) ([]VoyageReport, error)
	ByRiver(riverId int64, limitByGroup int) ([]VoyageReport, error)
	Insert(report VoyageReport) (int64, error)
	Update(report VoyageReport) error
	Find(id int64) (VoyageReport, bool, error)
	Remove(id int64) error
	UndoRemove(id int64) error

	UpsertVoyageReports(report ...VoyageReport) ([]VoyageReport, error)
	GetLastId(source string) (interface{}, error)
	AssociateWithRiver(voyageReportId, riverId int64) error
	ForEach(source string, callback func(report *VoyageReport) error) error
	RemoveRiverLink(id int64, tx interface{}) error
}

type SrtmDao interface {
	GetRaster(lat int, lon int) (Bytearea2D, bool, error)
}
