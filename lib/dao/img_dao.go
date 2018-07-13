package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/pkg/errors"
	"time"
	"fmt"
	log "github.com/Sirupsen/logrus"
)

type imgStorage struct {
	PostgresStorage
	upsertQuery      string
	listQuery        string
	insertLocalQuery string
	deleteQuery string
}

func NewImgPostgresDao(postgresStorage PostgresStorage) ImgDao {
	return imgStorage{
		PostgresStorage: postgresStorage,
		upsertQuery : queries.SqlQuery("img", "upsert"),
		listQuery : queries.SqlQuery("img", "list"),
		insertLocalQuery : queries.SqlQuery("img", "insert-local"),
		deleteQuery : queries.SqlQuery("img", "delete"),
	}
}

func (this imgStorage) Upsert(imgs ...Img) ([]Img, error) {

	ids, err := this.updateReturningId(this.upsertQuery,
		func(entity interface{}) ([]interface{}, error) {
			_img := entity.(Img)
			return []interface{}{_img.ReportId, _img.WwId, _img.Source, _img.RemoteId, _img.Url, _img.PreviewUrl, _img.DatePublished}, nil
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

func (this imgStorage) List(wwId int64, limit int) ([]Img, error) {
	result, err := this.doFindList(this.listQuery, func(rows *sql.Rows) (Img, error) {
		img := Img{}
		err := rows.Scan(&img.Id, &img.ReportId, &img.WwId, &img.Source, &img.RemoteId, &img.Url, &img.PreviewUrl, &img.DatePublished)
		if err != nil {
			return img, err
		}
		return img, nil
	}, wwId, limit)
	if err != nil {
		return []Img{}, err
	}
	return result.([]Img), nil
}

func (this imgStorage) InsertLocal(wwId int64, source string, urlBase string, previewUrlBase string, datePublished time.Time) (Img, error) {
	params := []interface{}{wwId, source, urlBase, previewUrlBase, datePublished}
	vals, err := this.updateReturningColumns(this.insertLocalQuery,
		func(entity interface{}) ([]interface{}, error) {
			return entity.([]interface{}), nil
		}, params)
	if err != nil {
		return Img{}, err
	}
	if len(vals) < 1 {
		return Img{}, errors.New("Image not inserted!")
	}
	row := vals[0]
	id := *row[0].(*int64)
	result := Img{
		Id:id,
		WwId:wwId,
		Source:source,
		RemoteId:fmt.Sprintf("%d", id),
		DatePublished:datePublished,
		Url:*row[1].(*string),
		PreviewUrl:*row[2].(*string),
	}

	fmt.Println(result)
	return result, nil
}

func (this imgStorage) Remove(id int64) error {
	log.Infof("Remove spot %d", id)
	return this.performUpdates(this.deleteQuery, idMapper, id)
}

func (this imgStorage) toInterface(imgs ...Img) []interface{} {
	imgs_i := make([]interface{}, len(imgs))
	for i := 0; i < len(imgs); i++ {
		imgs_i[i] = imgs[i]
	}
	return imgs_i
}