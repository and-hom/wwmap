package dao

import (
	"encoding/json"
	"fmt"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/lib/pq"
)

func NewTilePostgresDao(postgresStorage PostgresStorage) TileDao {
	return &tileStorage{
		PostgresStorage:   postgresStorage,
		insideBoundsQuery: queries.SqlQuery("tile", "inside-bounds"),
		singleRiverQuery:  queries.SqlQuery("tile", "by-id"),
	}
}

type tileStorage struct {
	PostgresStorage
	insideBoundsQuery string
	singleRiverQuery  string
}

func (this *tileStorage) GetRiver(riverId int64, imgLimit int) (RiverWithSpotsExt, error) {
	river := RiverWithSpotsExt{}

	rows, err := this.db.Query(this.singleRiverQuery, riverId,
		pq.Array([]string{string(IMAGE_TYPE_IMAGE), string(IMAGE_TYPE_VIDEO)}),
		imgLimit)
	if err != nil {
		return river, err
	}
	defer rows.Close()

	lastSpotId := int64(-1)

	for rows.Next() {
		spot := Spot{}
		img := Img{}

		pointStr := ""
		categoryStr := ""
		riverPropsStr := ""
		spotPropsStr := ""

		err := rows.Scan(&river.Id, &river.Title, &river.Description, &riverPropsStr, &river.Region.Id, &river.Region.Title, &river.Region.CountryId,
			&spot.Id, &spot.Title, &spot.Description, &pointStr, &categoryStr, &spot.Link, &spotPropsStr,
			&img.Id, &img.Source, &img.RemoteId, &img.Url, &img.PreviewUrl, &img.DatePublished, &img.Type)

		if err != nil {
			return river, err
		}

		if river.Props == nil {
			err = json.Unmarshal([]byte(riverPropsStr), &river.Props)
			if err != nil {
				return river, err
			}
		}

		if lastSpotId != spot.Id {
			err = json.Unmarshal(categoryStrBytes(categoryStr), &spot.Category)
			if err != nil {
				return river, err
			}

			pgPoint := PgPointOrLineString{}
			err = json.Unmarshal([]byte(pointStr), &pgPoint)
			if err != nil {
				return river, err
			}
			spot.Point = pgPoint.Coordinates

			err = json.Unmarshal([]byte(spotPropsStr), &spot.Props)
			if err != nil {
				return river, err
			}
			lastSpotId = spot.Id
		}

		lSp := len(river.Spots)
		if lSp == 0 || river.Spots[lSp-1].Id != spot.Id {
			river.Spots = append(river.Spots, spot)
			lSp += 1
		}

		if img.Id > 0 {
			river.Spots[lSp-1].Images = append(river.Spots[lSp-1].Images, img)
		}
	}

	if len(river.Spots) == 0 {
		// no records
		return river, fmt.Errorf("River with id %d not found or have no spots", riverId)
	}

	return river, nil
}

func (this *tileStorage) ListRiversWithBounds(bbox geo.Bbox, imgLimit int, showUnpublished bool) ([]RiverWithSpots, error) {
	rows, err := this.db.Query(this.insideBoundsQuery, bbox.Y1, bbox.X1, bbox.Y2, bbox.X2, imgLimit, showUnpublished)
	if err != nil {
		return []RiverWithSpots{}, err
	}
	defer rows.Close()

	rivers := make([]RiverWithSpots, 0)

	lastSpotId := int64(-1)

	for rows.Next() {
		river := RiverWithSpots{}
		spot := Spot{}
		img := Img{}

		pointStr := ""
		categoryStr := ""
		propsStr := ""

		err := rows.Scan(&river.Id, &river.Title, &river.RegionId, &river.CountryId,
			&spot.Id, &spot.Title, &spot.Description, &pointStr, &categoryStr, &spot.Link, &propsStr,
			&img.Id, &img.Source, &img.RemoteId, &img.Url, &img.PreviewUrl, &img.DatePublished, &img.Type)

		if err != nil {
			return []RiverWithSpots{}, err
		}

		if lastSpotId != spot.Id {
			err = json.Unmarshal(categoryStrBytes(categoryStr), &spot.Category)
			if err != nil {
				return []RiverWithSpots{}, err
			}

			pgPoint := PgPointOrLineString{}
			err = json.Unmarshal([]byte(pointStr), &pgPoint)
			if err != nil {
				return []RiverWithSpots{}, err
			}
			spot.Point = pgPoint.Coordinates.Flip()

			err = json.Unmarshal([]byte(propsStr), &spot.Props)
			if err != nil {
				return []RiverWithSpots{}, err
			}
			lastSpotId = spot.Id
		}

		lRiv := len(rivers)
		if lRiv == 0 || rivers[lRiv-1].Id != river.Id {
			rivers = append(rivers, river)
			lRiv += 1
		}

		lSp := len(rivers[lRiv-1].Spots)
		if lSp == 0 || rivers[lRiv-1].Spots[lSp-1].Id != spot.Id {
			rivers[lRiv-1].Spots = append(rivers[lRiv-1].Spots, spot)
			lSp += 1
		}

		if img.Id > 0 {
			rivers[lRiv-1].Spots[lSp-1].Images = append(rivers[lRiv-1].Spots[lSp-1].Images, img)
		}
	}

	return rivers, nil
}
