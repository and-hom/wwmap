package dao

import (
	"time"
	log "github.com/Sirupsen/logrus"
	"database/sql"
	_ "github.com/lib/pq"
	"errors"
	"reflect"
	. "github.com/and-hom/wwmap/lib/geo"
	"fmt"
	"github.com/and-hom/wwmap/lib/config"
)

type Storage interface {
	// Call payload function within transaction if supported by storage. Simply call payload function if not supported.
	WithinTx(payload func(tx interface{}) error) error
}

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

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(c config.Db) PostgresStorage {
	db, err := sql.Open("postgres", c.ConnString)
	if err != nil {
		log.Fatalf("Can not connect to postgres: %v", err)
	}
	db.SetConnMaxLifetime(c.MaxConnLifetime)
	db.SetMaxOpenConns(c.MaxOpenConn)
	db.SetMaxIdleConns(c.MaxIddleConn)

	return PostgresStorage{
		db:db,
	}
}

func NewPostgresStorageForDb(db *sql.DB) PostgresStorage {
	return PostgresStorage{db}
}

func nullIf0(x int64) sql.NullInt64 {
	if x == 0 {
		return sql.NullInt64{
			Int64:0,
			Valid:false,
		}
	}
	return sql.NullInt64{
		Int64:x,
		Valid:true,
	}
}

func getOrElse(val sql.NullInt64, _default int64) int64 {
	if val.Valid {
		return val.Int64
	} else {
		return _default
	}
}

func (this *PostgresStorage) doFindAndReturn(query string, callback interface{}, args ...interface{}) (interface{}, bool, error) {
	rows, err := this.db.Query(query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	funcValue := reflect.ValueOf(callback)

	for rows.Next() {
		val := funcValue.Call([]reflect.Value{reflect.ValueOf(rows)})
		if val[1].Interface() == nil {
			return val[0].Interface(), true, nil
		} else {
			return nil, false, val[1].Interface().(error)
		}
	}
	return nil, false, nil
}

func (this *PostgresStorage) doFindList(query string, callback interface{}, args ...interface{}) (interface{}, error) {
	rows, err := this.db.Query(query, args...)
	if err != nil {
		return []interface{}{}, err
	}
	defer rows.Close()

	funcValue := reflect.ValueOf(callback)
	returnType := funcValue.Type().Out(0)
	var result = reflect.MakeSlice(reflect.SliceOf(returnType), 0, 0)

	var lastErr error = nil
	for rows.Next() {
		val := funcValue.Call([]reflect.Value{reflect.ValueOf(rows)})
		if val[1].Interface() == nil {
			result = reflect.Append(result, val[0])
		} else {
			log.Error(val[1])
			lastErr = (val[1]).Interface().(error)
			break
		}
	}
	return result.Interface(), lastErr
}

func (this *PostgresStorage) forEach(query string, callback interface{}, args ...interface{}) error {
	rows, err := this.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	funcValue := reflect.ValueOf(callback)

	for rows.Next() {
		val := funcValue.Call([]reflect.Value{reflect.ValueOf(rows)})
		if val[0].Interface() != nil {
			return val[0].Interface().(error)
		}
	}
	return nil
}

// Deprecated: use updateReturningId
func (this *PostgresStorage) insertReturningId(query string, args ...interface{}) (int64, error) {
	tx, err := this.db.Begin()
	if err != nil {
		return -1, err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return -1, err
	}
	rows, err := stmt.Query(args...)
	if err != nil {
		return -1, err
	}

	lastId := int64(-1)
	for rows.Next() {
		rows.Scan(&lastId)
	}

	err = rows.Close()
	if err != nil {
		return -1, err
	}
	err = stmt.Close()
	if err != nil {
		return -1, err
	}
	err = tx.Commit()
	if err != nil {
		return -1, err
	}
	if lastId < 0 {
		return -1, errors.New("Not inserted")
	}
	return lastId, nil
}

func (this *PostgresStorage) updateReturningId(query string, mapper func(entity interface{}) ([]interface{}, error), values ...interface{}) ([]int64, error) {
	rows, err := this.updateReturningColumns(query, mapper, values...)
	if err != nil {
		return []int64{}, err
	}
	result := make([]int64, len(rows))
	for i, row := range rows {
		result[i] = *row[0].(*int64)
	}
	return result, nil
}

func (this *PostgresStorage) updateReturningColumns(query string, mapper func(entity interface{}) ([]interface{}, error), values ...interface{}) ([][]interface{}, error) {
	tx, err := this.db.Begin()
	if err != nil {
		return [][]interface{}{}, err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return [][]interface{}{}, err
	}

	result := make([][]interface{}, len(values))
	for idx, value := range values {
		args, err := mapper(value)
		if err != nil {
			return [][]interface{}{}, err
		}
		rows, err := stmt.Query(args...)
		if err != nil {
			return [][]interface{}{}, err
		}
		colTypes, err := rows.ColumnTypes()
		if err != nil {
			return [][]interface{}{}, err
		}
		if rows.Next() {
			result[idx] = make([]interface{}, len(colTypes))
			for i, t := range colTypes {
				result[idx][i] = reflect.New(t.ScanType()).Interface()
			}
			rows.Scan(result[idx]...)
		} else {
			return [][]interface{}{}, fmt.Errorf("Value is not inserted: %v+\n %s", args, query)
		}
		err = rows.Close()
		if err != nil {
			return [][]interface{}{}, err
		}
	}

	err = stmt.Close()
	if err != nil {
		return [][]interface{}{}, err
	}
	err = tx.Commit()
	if err != nil {
		return [][]interface{}{}, err
	}
	return result, nil
}

func (this *PostgresStorage) performUpdates(query string, mapper func(entity interface{}) ([]interface{}, error), values ...interface{}) error {
	return this.WithinTx(func(tx interface{}) error {
		txHolder := tx.(PgTxHolder)
		return (&txHolder).performUpdates(query, mapper, values...)
	})
}

type PgPoint struct {
	Coordinates Point `json:"coordinates"`
}

func (this PgPoint) GetPoint() Point {
	// flip coordinates for postGIS
	return this.Coordinates.Flip()
}

type PgPolygon struct {
	Coordinates [][]Point `json:"coordinates"`
}


