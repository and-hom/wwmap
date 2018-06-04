package dao

type ImgStorage struct {
	PostgresStorage
}

func (this ImgStorage) Upsert(imgs ...Img) ([]Img, error) {
	imgs_i := make([]interface{}, len(imgs))
	for i := 0; i < len(imgs); i++ {
		imgs_i[i] = imgs[i]
	}
	ids, err := this.updateReturningId("INSERT INTO image(white_water_rapid_id,source,remote_id,url,preview_url,date_published) " +
		"VALUES ($1, $2, $3, $4, $5, $6) " +
		"ON CONFLICT DO NOTHING " +
		"RETURNING id",
		func(entity interface{}) ([]interface{}, error) {
			_report := entity.(Img)
			return []interface{}{_report.WwId, _report.Source, _report.RemoteId, _report.Url, _report.PreviewUrl, _report.DatePublished}, nil
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