package dao

import (
	log "github.com/Sirupsen/logrus"
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/model"
	"fmt"
	"strings"
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewWhiteWaterPostgresDao(postgresStorage PostgresStorage) WhiteWaterDao {
	return whiteWaterStorage{
		PostgresStorage:postgresStorage,
		listByBoxQuery: queries.SqlQuery("white-water", "by-box"),
		listByRiverQuery: queries.SqlQuery("white-water", "by-river"),
		listByRiverAndTitleQuery: queries.SqlQuery("white-water", "by-river-and-title"),
		listWithPathQuery: queries.SqlQuery("white-water", "with-path"),
		insertQuery: queries.SqlQuery("white-water", "insert"),
	}
}

type whiteWaterStorage struct {
	PostgresStorage
	listByBoxQuery           string
	listByRiverQuery         string
	listByRiverAndTitleQuery string
	listWithPathQuery        string
	insertQuery              string
}

func (this whiteWaterStorage) ListWithPath() ([]WhiteWaterPointWithPath, error) {
	return this.listWithPath("");
}

func (this whiteWaterStorage) ListByBbox(bbox geo.Bbox) ([]WhiteWaterPointWithRiverTitle, error) {
	return this.list(this.listByBoxQuery, bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (this whiteWaterStorage) ListByRiver(id int64) ([]WhiteWaterPointWithRiverTitle, error) {
	return this.list(this.listByRiverQuery, id)
}

func (this whiteWaterStorage) ListByRiverAndTitle(riverId int64, title string) ([]WhiteWaterPointWithRiverTitle, error) {
	return this.list(this.listByRiverAndTitleQuery, riverId, title)
}

func (this whiteWaterStorage) list(query string, vars ...interface{}) ([]WhiteWaterPointWithRiverTitle, error) {
	result, err := this.doFindList(query,
		func(rows *sql.Rows) (WhiteWaterPointWithRiverTitle, error) {

			riverTitle := sql.NullString{}

			wwPoint, err := scanWwPoint(rows, &riverTitle)
			if err != nil {
				return WhiteWaterPointWithRiverTitle{}, err
			}

			whiteWaterPoint := WhiteWaterPointWithRiverTitle{
				wwPoint,
				riverTitle.String,
				[]Img{},
			}
			return whiteWaterPoint, nil
		}, vars...)
	if (err != nil ) {
		return []WhiteWaterPointWithRiverTitle{}, err
	}
	return result.([]WhiteWaterPointWithRiverTitle), nil
}

func (this whiteWaterStorage) listWithPath(query string, vars ...interface{}) ([]WhiteWaterPointWithPath, error) {
	result, err := this.doFindList(query,
		func(rows *sql.Rows) (WhiteWaterPointWithPath, error) {

			riverTitle := ""
			regionTitle := sql.NullString{}
			countryTitle := ""

			wwPoint, err := scanWwPoint(rows, &riverTitle, &regionTitle, &countryTitle)
			if err != nil {
				return WhiteWaterPointWithPath{}, err
			}

			path := []string{}
			if regionTitle.String == "" {
				path = []string{countryTitle, riverTitle, wwPoint.Title}
			} else {
				path = []string{countryTitle, regionTitle.String, riverTitle, wwPoint.Title}
			}

			whiteWaterPoint := WhiteWaterPointWithPath{
				wwPoint,
				path,
			}
			return whiteWaterPoint, nil
		}, vars...)
	if (err != nil ) {
		return []WhiteWaterPointWithPath{}, err
	}
	return result.([]WhiteWaterPointWithPath), nil
}

func scanWwPoint(rows *sql.Rows, additionalVars ...interface{}) (WhiteWaterPoint, error) {
	var err error
	id := int64(-1)
	osmId := sql.NullInt64{}
	_type := ""
	title := ""
	comment := ""
	pointStr := ""
	categoryStr := ""
	shortDesc := sql.NullString{}
	link := sql.NullString{}
	riverId := sql.NullInt64{}

	fields := append([]interface{}{&id, &osmId, &_type, &title, &comment, &pointStr, &categoryStr, &shortDesc, &link, &riverId}, additionalVars...)
	err = rows.Scan(fields...)
	if err != nil {
		log.Errorf("Can not read from db: %v", err)
		return WhiteWaterPoint{}, err
	}

	var pgPoint PgPoint
	err = json.Unmarshal([]byte(pointStr), &pgPoint)
	if err != nil {
		log.Errorf("Can not parse point %s for white water object %d: %v", pointStr, id, err)
		return WhiteWaterPoint{}, err
	}

	category := model.SportCategory{}
	if !(strings.HasPrefix(categoryStr, "\"") && strings.HasSuffix(categoryStr, "\"")) {
		categoryStr = "\"" + categoryStr + "\""
	}
	err = json.Unmarshal([]byte(categoryStr), &category)
	if err != nil {
		log.Errorf("Can not parse category %s for white water object %d: %v", categoryStr, id, err)
		return WhiteWaterPoint{}, err
	}

	return WhiteWaterPoint{
		Id:id,
		OsmId:getOrElse(osmId, -1),
		RiverId:getOrElse(riverId, -1),
		Title: title,
		Type: _type,
		Point: pgPoint.Coordinates,
		Comment: comment,
		Category: category,
		ShortDesc: shortDesc.String,
		Link: link.String,
	}, nil
}

func (this whiteWaterStorage) AddWhiteWaterPoints(whiteWaterPoints ...WhiteWaterPoint) error {
	vars := make([]interface{}, len(whiteWaterPoints))
	for i, p := range whiteWaterPoints {
		vars[i] = p
	}
	return this.performUpdates(this.insertQuery,
		func(entity interface{}) ([]interface{}, error) {
			wwp := entity.(WhiteWaterPoint)
			pathBytes, err := json.Marshal(geo.NewGeoPoint(wwp.Point))
			if err != nil {
				return nil, err
			}
			fmt.Printf("id = %d", wwp.Id)
			cat, err := wwp.Category.MarshalJSON()
			if err != nil {
				return nil, err
			}
			return []interface{}{wwp.OsmId, wwp.Title, wwp.Type, cat, wwp.Comment, string(pathBytes), wwp.ShortDesc, wwp.Link, nullIf0(wwp.RiverId)}, nil
		}, vars...)
}

