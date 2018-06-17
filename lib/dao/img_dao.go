package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
)

type imgStorage struct {
	PostgresStorage
	upsertQuery string
	listQuery   string
}

func NewImgPostgresDao(postgresStorage PostgresStorage) ImgDao {
	return imgStorage{
		PostgresStorage: postgresStorage,
		upsertQuery : queries.SqlQuery("img","upsert"),
		listQuery : queries.SqlQuery("img","list"),
	}
}

func (this imgStorage) Upsert(imgs ...Img) ([]Img, error) {
	imgs_i := make([]interface{}, len(imgs))
	for i := 0; i < len(imgs); i++ {
		imgs_i[i] = imgs[i]
	}
	ids, err := this.updateReturningId(this.upsertQuery,
		func(entity interface{}) ([]interface{}, error) {
			_img := entity.(Img)
			return []interface{}{_img.ReportId, _img.WwId, _img.Source, _img.RemoteId, _img.Url, _img.PreviewUrl, _img.DatePublished}, nil
		}, imgs_i...)

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