package dao

import (
	"database/sql"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"html"
	"time"
)

var zeroDate = util.ZeroDateUTC()

type imgStorage struct {
	PostgresStorage
	upsertQuery             string
	findQuery               string
	listQuery               string
	listExtQuery            string
	listAllBySpotQuery      string
	listAllByRiverQuery     string
	listMainByRiverQuery    string
	insertLocalQuery        string
	deleteQuery             string
	setEnabledQuery         string
	getMainForSpotQuery     string
	setMainQuery            string
	dropMainForSpotQuery    string
	setSetLevelAndDateQuery string
	setManualLevelQuery     string
	resetManualLevelQuery   string
	deleteForSpot           string
	deleteForRiver          string
	parentIds               string
}

func NewImgPostgresDao(postgresStorage PostgresStorage) ImgDao {
	return imgStorage{
		PostgresStorage:         postgresStorage,
		upsertQuery:             queries.SqlQuery("img", "upsert"),
		findQuery:               queries.SqlQuery("img", "by-id"),
		listQuery:               queries.SqlQuery("img", "list"),
		listExtQuery:            queries.SqlQuery("img", "list-ext"),
		listAllBySpotQuery:      queries.SqlQuery("img", "list-all-by-spot"),
		listAllByRiverQuery:     queries.SqlQuery("img", "list-all-by-river"),
		listMainByRiverQuery:    queries.SqlQuery("img", "list-main-by-river"),
		insertLocalQuery:        queries.SqlQuery("img", "insert-local"),
		deleteQuery:             queries.SqlQuery("img", "delete"),
		setEnabledQuery:         queries.SqlQuery("img", "set-enabled"),
		getMainForSpotQuery:     queries.SqlQuery("img", "get-main"),
		setMainQuery:            queries.SqlQuery("img", "set-main"),
		dropMainForSpotQuery:    queries.SqlQuery("img", "drop-main-for-spot"),
		setSetLevelAndDateQuery: queries.SqlQuery("img", "set-level-and-date"),
		setManualLevelQuery:     queries.SqlQuery("img", "set-manual-level"),
		resetManualLevelQuery:   queries.SqlQuery("img", "reset-manual-level"),
		deleteForSpot:           queries.SqlQuery("img", "delete-by-spot"),
		deleteForRiver:          queries.SqlQuery("img", "delete-by-river"),
		parentIds:               queries.SqlQuery("img", "parent-ids"),
	}
}

func (this imgStorage) Upsert(imgs ...Img) ([]Img, error) {

	ids, err := this.UpdateReturningId(this.upsertQuery, func(entity interface{}) ([]interface{}, error) {
		_img := entity.(Img)
		return []interface{}{_img.ReportId, _img.WwId, _img.Source, _img.RemoteId, _img.Url, _img.PreviewUrl, _img.DatePublished, _img.Type}, nil
	}, true, this.toInterface(imgs...)...)

	if err != nil {
		return []Img{}, err
	}

	result := make([]Img, len(imgs))
	copy(result, imgs)
	for i := 0; i < len(imgs); i++ {
		result[i].Id = ids[i]
	}
	return result, nil
}

func imgMapper(rows *sql.Rows) (Img, error) {
	img := Img{}
	var levelString sql.NullString
	var dateLevelUpdated pq.NullTime
	var date pq.NullTime
	err := rows.Scan(&img.Id, &img.ReportId, &img.WwId, &img.Source, &img.RemoteId, &img.Url, &img.PreviewUrl,
		&img.DatePublished, &img.Enabled, &img.Type, &img.MainImage, &date, &dateLevelUpdated, &levelString)
	if err != nil {
		return img, err
	}
	if levelString.Valid {
		err = json.Unmarshal([]byte(levelString.String), &img.Level)
		if err != nil {
			return img, err
		}
	}
	img.DateLevelUpdated = nullDateToZero(dateLevelUpdated)
	img.Date = nullDateToPtr(date)
	return img, nil
}

func imgExtMapper(rows *sql.Rows) (ImgExt, error) {
	img := ImgExt{}
	var levelString sql.NullString
	var dateLevelUpdated pq.NullTime
	var date pq.NullTime
	err := rows.Scan(&img.Id, &img.ReportId, &img.WwId, &img.Source, &img.RemoteId, &img.Url, &img.PreviewUrl,
		&img.DatePublished, &img.Enabled, &img.Type, &img.MainImage, &date, &dateLevelUpdated, &levelString,
		&img.ReportUrl, &img.ReportTitle)
	if err != nil {
		return img, err
	}
	if levelString.Valid {
		err = json.Unmarshal([]byte(levelString.String), &img.Level)
		if err != nil {
			return img, err
		}
	}
	img.DateLevelUpdated = nullDateToZero(dateLevelUpdated)
	img.Date = nullDateToPtr(date)
	// workaround: unescape report title on store to database
	img.ReportTitle = html.UnescapeString(img.ReportTitle)
	return img, nil
}

func (this imgStorage) List(wwId int64, limit int, _type ImageType, enabledOnly bool) ([]Img, error) {
	result, err := this.DoFindList(this.listQuery, imgMapper, wwId, _type, enabledOnly, limit)
	if err != nil {
		return []Img{}, err
	}
	return result.([]Img), nil
}

func (this imgStorage) ListExt(wwId int64, limit int, _type ImageType, enabledOnly bool) ([]ImgExt, error) {
	result, err := this.DoFindList(this.listExtQuery, imgExtMapper, wwId, _type, enabledOnly, limit)
	if err != nil {
		return []ImgExt{}, err
	}
	return result.([]ImgExt), nil
}

func (this imgStorage) ListAllBySpot(wwId int64) ([]Img, error) {
	result, err := this.DoFindList(this.listAllBySpotQuery, imgMapper, wwId)
	if err != nil {
		return []Img{}, err
	}
	return result.([]Img), nil
}

func (this imgStorage) ListAllByRiver(riverId int64) ([]Img, error) {
	result, err := this.DoFindList(this.listAllByRiverQuery, imgMapper, riverId)
	if err != nil {
		return []Img{}, err
	}
	return result.([]Img), nil
}

func (this imgStorage) ListMainByRiver(riverId int64) ([]Img, error) {
	result, err := this.DoFindList(this.listMainByRiverQuery, imgMapper, riverId, string(IMAGE_TYPE_IMAGE))
	if err != nil {
		return []Img{}, err
	}
	return result.([]Img), nil
}

func (this imgStorage) Find(id int64) (Img, bool, error) {
	result, found, err := this.DoFindAndReturn(this.findQuery, imgMapper, id)
	if err != nil {
		return Img{}, found, err
	}
	return result.(Img), found, nil
}

func (this imgStorage) GetMainForSpot(spotId int64) (Img, bool, error) {
	result, found, err := this.DoFindAndReturn(this.getMainForSpotQuery, imgMapper, spotId)
	if err != nil || !found {
		return Img{}, found, err
	}
	return result.(Img), found, nil
}

func (this imgStorage) InsertLocal(wwId int64, _type ImageType, source string, urlBase string, previewUrlBase string, datePublished time.Time) (Img, error) {
	params := []interface{}{wwId, _type, source, datePublished}
	vals, err := this.UpdateReturningColumns(this.insertLocalQuery, ArrayMapper, true, params)
	if err != nil {
		return Img{}, err
	}
	if len(vals) < 1 {
		return Img{}, errors.New("Image not inserted!")
	}
	row := vals[0]
	id := *row[0].(*int64)
	enabled := *row[1].(*bool)
	result := Img{
		Id:            id,
		WwId:          wwId,
		Source:        source,
		RemoteId:      fmt.Sprintf("%d", id),
		DatePublished: datePublished,
		Url:           "",
		PreviewUrl:    "",
		Type:          _type,
		Enabled:       enabled,
	}

	return result, nil
}

func (this imgStorage) Remove(id int64, tx interface{}) error {
	log.Infof("Remove image %d", id)
	return this.PerformUpdatesWithinTxOptionally(tx, this.deleteQuery, IdMapper, id)
}

func (this imgStorage) SetEnabled(id int64, enabled bool) error {
	return this.PerformUpdates(this.setEnabledQuery, ArrayMapper, []interface{}{enabled, id})
}

func (this imgStorage) toInterface(imgs ...Img) []interface{} {
	imgs_i := make([]interface{}, len(imgs))
	for i := 0; i < len(imgs); i++ {
		imgs_i[i] = imgs[i]
	}
	return imgs_i
}

func (this imgStorage) SetMain(spotId int64, id int64) error {
	return this.PerformUpdates(this.setMainQuery, ArrayMapper, []interface{}{spotId, id})
}

func (this imgStorage) DropMainForSpot(spotId int64) error {
	return this.PerformUpdates(this.dropMainForSpotQuery, IdMapper, spotId)
}

func (this imgStorage) SetDateAndLevel(id int64, date time.Time, level map[string]int8) error {
	var nullableDate pq.NullTime
	if date == zeroDate {
		nullableDate = pq.NullTime{Valid: false}
	} else {
		nullableDate = pq.NullTime{Time: date, Valid: true}
	}

	levelB, err := json.Marshal(level)
	if err != nil {
		return err
	}

	return this.PerformUpdates(this.setSetLevelAndDateQuery, ArrayMapper, []interface{}{id, nullableDate, string(levelB)})
}

func (this imgStorage) SetManualLevel(id int64, level int8) (map[string]int8, error) {
	vals, err := this.UpdateReturningColumns(this.setManualLevelQuery, ArrayMapper, true, []interface{}{id, int(level)})
	if err != nil {
		return nil, err
	}
	var levels map[string]int8
	levelB := []byte(*(vals[0][0].(*string)))
	err = json.Unmarshal(levelB, &levels)
	if err != nil {
		return nil, err
	}
	return levels, nil
}

func (this imgStorage) ResetManualLevel(id int64) (map[string]int8, error) {
	vals, err := this.UpdateReturningColumns(this.resetManualLevelQuery, IdMapper, true, id)
	if err != nil {
		return nil, err
	}
	var levels map[string]int8
	levelB := []byte(*(vals[0][0].(*string)))
	err = json.Unmarshal(levelB, &levels)
	if err != nil {
		return nil, err
	}
	return levels, nil
}

func (this imgStorage) RemoveBySpot(spotId int64, tx interface{}) error {
	return this.PerformUpdatesWithinTxOptionally(tx, this.deleteForSpot, IdMapper, spotId)
}

func (this imgStorage) RemoveByRiver(riverId int64, tx interface{}) error {
	return this.PerformUpdatesWithinTxOptionally(tx, this.deleteForRiver, IdMapper, riverId)
}

func (this imgStorage) GetParentIds(imgIds []int64) (map[int64]ImageParentIds, error) {
	result := make(map[int64]ImageParentIds)

	_, err := this.DoFindList(this.parentIds, func(rows *sql.Rows) (int, error) {
		imgId := int64(0)
		parentIds := ImageParentIds{}
		err := rows.Scan(&imgId, &parentIds.SpotId)

		if err == nil {
			result[imgId] = parentIds
		}
		return 0, err
	}, pq.Array(imgIds))

	if err != nil {
		return result, err
	}
	return result, nil
}
