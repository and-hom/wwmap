package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/geo"
	"encoding/json"
)

func NewTilePostgresDao(postgresStorage PostgresStorage) TileDao {
	return &tileStorage{
		PostgresStorage: postgresStorage,
		insideBoundsQuery: queries.SqlQuery("tile", "inside-bounds"),
	}
}

type tileStorage struct {
	PostgresStorage
	insideBoundsQuery string
}

func (this *tileStorage) ListRiversWithBounds(bbox geo.Bbox, showUnpublished bool, imgLimit int) ([]RiverWithSpots, error) {
	rows, err := this.db.Query(this.insideBoundsQuery, bbox.Y1, bbox.X1, bbox.Y2, bbox.X2, imgLimit, showUnpublished)
	if err != nil {
		return []RiverWithSpots{}, err
	}
	defer rows.Close()

	rivers := make([]RiverWithSpots, 0)

	river := RiverWithSpots{}
	spot := Spot{}
	img := Img{}

	lastSpotId := int64(-1)

	for rows.Next() {

		pointStr := ""
		categoryStr := ""
		propsStr := ""

		err := rows.Scan(&river.Id, &river.Title,
			&spot.Id, &spot.Title, &pointStr, &categoryStr, &spot.Link, &propsStr,
			&img.Id, &img.Source, &img.RemoteId, &img.Url, &img.PreviewUrl, &img.DatePublished, &img.Type)

		if err != nil {
			return []RiverWithSpots{}, err
		}

		if lastSpotId != spot.Id {
			err = json.Unmarshal(categoryStrBytes(categoryStr), &spot.Category)
			if err != nil {
				return []RiverWithSpots{}, err
			}

			pgPoint := PgPoint{}
			err = json.Unmarshal([]byte(pointStr), &pgPoint)
			if err != nil {
				return []RiverWithSpots{}, err
			}
			spot.Point = pgPoint.GetPoint()

			err = json.Unmarshal([]byte(propsStr), &spot.Props)
			if err != nil {
				return []RiverWithSpots{}, err
			}
			lastSpotId = spot.Id
		}

		lRiv := len(rivers)
		if lRiv == 0 || rivers[lRiv - 1].Id != river.Id {
			rivers = append(rivers, river)
			lRiv += 1
		}

		lSp := len(rivers[lRiv - 1].Spots)
		if lSp == 0 || rivers[lRiv - 1].Spots[lSp - 1].Id != spot.Id {
			rivers[lRiv - 1].Spots = append(rivers[lRiv - 1].Spots, spot)
			lSp += 1
		}

		if img.Id > 0 {
			rivers[lRiv - 1].Spots[lSp - 1].Images = append(rivers[lRiv - 1].Spots[lSp - 1].Images, img)
		}
	}

	return rivers, nil
}
