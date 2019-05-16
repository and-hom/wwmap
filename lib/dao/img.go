package dao

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/pkg/errors"
	"time"
)

type imgStorage struct {
	PostgresStorage
	upsertQuery          string
	findQuery            string
	listQuery            string
	listAllBySpotQuery   string
	listAllByRiverQuery  string
	listMainByRiverQuery string
	insertLocalQuery     string
	deleteQuery          string
	setEnabledQuery      string
	getMainForSpotQuery  string
	setMainQuery         string
	dropMainForSpotQuery string
	deleteForSpot        string
	deleteForRiver       string
}

func NewImgPostgresDao(postgresStorage PostgresStorage) ImgDao {
	return imgStorage{
		PostgresStorage:      postgresStorage,
		upsertQuery:          queries.SqlQuery("img", "upsert"),
		findQuery:            queries.SqlQuery("img", "by-id"),
		listQuery:            queries.SqlQuery("img", "list"),
		listAllBySpotQuery:   queries.SqlQuery("img", "list-all-by-spot"),
		listAllByRiverQuery:  queries.SqlQuery("img", "list-all-by-river"),
		listMainByRiverQuery: queries.SqlQuery("img", "list-main-by-river"),
		insertLocalQuery:     queries.SqlQuery("img", "insert-local"),
		deleteQuery:          queries.SqlQuery("img", "delete"),
		setEnabledQuery:      queries.SqlQuery("img", "set-enabled"),
		getMainForSpotQuery:  queries.SqlQuery("img", "get-main"),
		setMainQuery:         queries.SqlQuery("img", "set-main"),
		dropMainForSpotQuery: queries.SqlQuery("img", "drop-main-for-spot"),
		deleteForSpot:        queries.SqlQuery("img", "delete-by-spot"),
		deleteForRiver:       queries.SqlQuery("img", "delete-by-river"),
	}
}

func (this imgStorage) Upsert(imgs ...Img) ([]Img, error) {

	ids, err := this.updateReturningId(this.upsertQuery,
		func(entity interface{}) ([]interface{}, error) {
			_img := entity.(Img)
			return []interface{}{_img.ReportId, _img.WwId, _img.Source, _img.RemoteId, _img.Url, _img.PreviewUrl, _img.DatePublished, _img.Type}, nil
		}, this.toInterface(imgs...)...)

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
	err := rows.Scan(&img.Id, &img.ReportId, &img.WwId, &img.Source, &img.RemoteId, &img.Url, &img.PreviewUrl,
		&img.DatePublished, &img.Enabled, &img.Type, &img.MainImage)
	if err != nil {
		return img, err
	}
	return img, nil
}

func (this imgStorage) List(wwId int64, limit int, _type ImageType, enabledOnly bool) ([]Img, error) {
	result, err := this.doFindList(this.listQuery, imgMapper, wwId, _type, enabledOnly, limit)
	if err != nil {
		return []Img{}, err
	}
	return result.([]Img), nil
}

func (this imgStorage) ListAllBySpot(wwId int64) ([]Img, error) {
	result, err := this.doFindList(this.listAllBySpotQuery, imgMapper, wwId)
	if err != nil {
		return []Img{}, err
	}
	return result.([]Img), nil
}

func (this imgStorage) ListAllByRiver(riverId int64) ([]Img, error) {
	result, err := this.doFindList(this.listAllByRiverQuery, imgMapper, riverId)
	if err != nil {
		return []Img{}, err
	}
	return result.([]Img), nil
}

func (this imgStorage) ListMainByRiver(riverId int64) ([]Img, error) {
	result, err := this.doFindList(this.listMainByRiverQuery, imgMapper, riverId, string(IMAGE_TYPE_IMAGE))
	if err != nil {
		return []Img{}, err
	}
	return result.([]Img), nil
}

func (this imgStorage) Find(id int64) (Img, bool, error) {
	result, found, err := this.doFindAndReturn(this.findQuery, imgMapper, id)
	if err != nil {
		return Img{}, found, err
	}
	return result.(Img), found, nil
}

func (this imgStorage) GetMainForSpot(spotId int64) (Img, bool, error) {
	result, found, err := this.doFindAndReturn(this.getMainForSpotQuery, imgMapper, spotId)
	if err != nil || !found {
		return Img{}, found, err
	}
	return result.(Img), found, nil
}

func (this imgStorage) InsertLocal(wwId int64, _type ImageType, source string, urlBase string, previewUrlBase string, datePublished time.Time) (Img, error) {
	params := []interface{}{wwId, _type, source, datePublished}
	vals, err := this.updateReturningColumns(this.insertLocalQuery, ArrayMapper, params)
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
	return this.performUpdatesWithinTxOptionally(tx, this.deleteQuery, IdMapper, id)
}

func (this imgStorage) SetEnabled(id int64, enabled bool) error {
	return this.performUpdates(this.setEnabledQuery, ArrayMapper, []interface{}{enabled, id})
}

func (this imgStorage) toInterface(imgs ...Img) []interface{} {
	imgs_i := make([]interface{}, len(imgs))
	for i := 0; i < len(imgs); i++ {
		imgs_i[i] = imgs[i]
	}
	return imgs_i
}

func (this imgStorage) SetMain(spotId int64, id int64) error {
	return this.performUpdates(this.setMainQuery, ArrayMapper, []interface{}{spotId, id})
}

func (this imgStorage) DropMainForSpot(spotId int64) error {
	return this.performUpdates(this.dropMainForSpotQuery, IdMapper, spotId)
}

func (this imgStorage) RemoveBySpot(spotId int64, tx interface{}) error {
	return this.performUpdatesWithinTxOptionally(tx, this.deleteForSpot, IdMapper, spotId)
}

func (this imgStorage) RemoveByRiver(riverId int64, tx interface{}) error {
	return this.performUpdatesWithinTxOptionally(tx, this.deleteForRiver, IdMapper, riverId)
}
