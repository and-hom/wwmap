package dao

import (
	log "github.com/Sirupsen/logrus"
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/model"
	"strings"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"fmt"
	"github.com/lib/pq"
	"github.com/and-hom/wwmap/lib/util"
)

func NewWhiteWaterPostgresDao(postgresStorage PostgresStorage) WhiteWaterDao {
	return whiteWaterStorage{
		PostgresStorage:postgresStorage,
		PropsManager:PropertyManagerImpl{table:queries.SqlQuery("white-water", "table"), dao:&postgresStorage},
		listByBoxQuery: queries.SqlQuery("white-water", "by-box"),
		listByRiverQuery: queries.SqlQuery("white-water", "by-river"),
		listByRiverFullQuery: queries.SqlQuery("white-water", "by-river-full"),
		listByRiverAndTitleQuery: queries.SqlQuery("white-water", "by-river-and-title"),
		listWithPathQuery: queries.SqlQuery("white-water", "with-path"),
		insertQuery: queries.SqlQuery("white-water", "insert"),
		insertFullQuery: queries.SqlQuery("white-water", "insert-full"),
		updateQuery: queries.SqlQuery("white-water", "update"),
		byIdQuery: queries.SqlQuery("white-water", "by-id"),
		byIdFullQuery: queries.SqlQuery("white-water", "by-id-full"),
		updateFullQuery: queries.SqlQuery("white-water", "update-full"),
		deleteQuery: queries.SqlQuery("white-water", "delete"),
		deleteForRiverQuery: queries.SqlQuery("white-water", "delete-for-river"),
		geomCenterByRiverQuery: queries.SqlQuery("white-water", "geom-center-by-river"),
	}
}

type whiteWaterStorage struct {
	PostgresStorage
	PropsManager             PropertyManager
	listByBoxQuery           string
	listByRiverQuery         string
	listByRiverFullQuery     string
	listByRiverAndTitleQuery string
	listWithPathQuery        string
	insertQuery              string
	insertFullQuery          string
	updateQuery              string
	byIdQuery                string
	byIdFullQuery            string
	updateFullQuery          string
	deleteQuery              string
	deleteForRiverQuery              string
	geomCenterByRiverQuery   string
}

func (this whiteWaterStorage) ListWithPath() ([]WhiteWaterPointWithPath, error) {
	return this.listWithPath(this.listWithPathQuery);
}

func (this whiteWaterStorage) ListByBbox(bbox geo.Bbox) ([]WhiteWaterPointWithRiverTitle, error) {
	return this.list(this.listByBoxQuery, bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (this whiteWaterStorage) ListByRiver(id int64) ([]WhiteWaterPointWithRiverTitle, error) {
	return this.list(this.listByRiverQuery, id)
}

func (this whiteWaterStorage) ListByRiverFull(id int64) ([]WhiteWaterPointFull, error) {
	return this.listFull(this.listByRiverFullQuery, id)
}

func (this whiteWaterStorage) ListByRiverAndTitle(riverId int64, title string) ([]WhiteWaterPointWithRiverTitle, error) {
	return this.list(this.listByRiverAndTitleQuery, riverId, title)
}

func (this whiteWaterStorage) Find(id int64) (WhiteWaterPointWithRiverTitle, error) {
	found, err := this.list(this.byIdQuery, id)
	if err != nil {
		return WhiteWaterPointWithRiverTitle{}, err
	}
	if len(found) == 0 {
		return WhiteWaterPointWithRiverTitle{}, fmt.Errorf("Spot with id %d not found", id)
	}
	return found[0], nil
}

func (this whiteWaterStorage) FindFull(id int64) (WhiteWaterPointFull, error) {
	result, found, err := this.doFindAndReturn(this.byIdFullQuery, func(rows *sql.Rows) (interface{}, error) {
		return scanWwPointFull(rows)
	}, id)
	if err != nil {
		return WhiteWaterPointFull{}, err
	}
	if !found {
		return WhiteWaterPointFull{}, fmt.Errorf("Spot with id %d not found", id)
	}
	return result.(WhiteWaterPointFull), nil
}

func (this whiteWaterStorage) InsertWhiteWaterPointFull(whiteWaterPoint WhiteWaterPointFull) (int64, error) {
	params, err := paramsFull(whiteWaterPoint)
	if err != nil {
		return -1, nil
	}
	fmt.Println(params[15])
	fmt.Println(params[16])
	return this.insertReturningId(this.insertFullQuery, params...)
}

func (this whiteWaterStorage) UpdateWhiteWaterPointsFull(whiteWaterPoints ...WhiteWaterPointFull) error {
	vars := make([]interface{}, len(whiteWaterPoints))
	for i, p := range whiteWaterPoints {
		vars[i] = p
	}
	return this.performUpdates(this.updateFullQuery,
		func(entity interface{}) ([]interface{}, error) {
			wwp := entity.(WhiteWaterPointFull)
			params, err := paramsFull(wwp)
			if err != nil {
				return nil, err
			}
			return append([]interface{}{wwp.Id}, params...), nil
		}, vars...)
}

func paramsFull(wwp WhiteWaterPointFull) ([]interface{}, error) {
	pointBytes, err := json.Marshal(geo.NewGeoPoint(wwp.Point))
	if err != nil {
		return nil, err
	}

	cat, err := wwp.Category.MarshalJSON()
	if err != nil {
		return nil, err
	}
	lwCat, err := wwp.LowWaterCategory.MarshalJSON()
	if err != nil {
		return nil, err
	}
	mwCat, err := wwp.MediumWaterCategory.MarshalJSON()
	if err != nil {
		return nil, err
	}
	hwCat, err := wwp.HighWaterCategory.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return []interface{}{wwp.Title, string(cat), string(pointBytes), wwp.ShortDesc, wwp.Link, nullIf0(wwp.River.Id),
		string(lwCat), wwp.LowWaterDescription, string(mwCat), wwp.MediumWaterDescription, string(hwCat), wwp.HighWaterDescription,
		wwp.Orient, wwp.Approach, wwp.Safety, wwp.OrderIndex, wwp.AutomaticOrdering}, nil
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

func (this whiteWaterStorage) listFull(query string, vars ...interface{}) ([]WhiteWaterPointFull, error) {
	result, err := this.doFindList(query, scanWwPointFull, vars...)
	if (err != nil ) {
		return []WhiteWaterPointFull{}, err
	}
	return result.([]WhiteWaterPointFull), nil
}

func (this whiteWaterStorage) listWithPath(query string, vars ...interface{}) ([]WhiteWaterPointWithPath, error) {
	result, err := this.doFindList(query,
		func(rows *sql.Rows) (WhiteWaterPointWithPath, error) {

			regionTitle := sql.NullString{}
			countryTitle := ""

			wwPoint, err := scanWwPointFull(rows, &regionTitle, &countryTitle)
			if err != nil {
				return WhiteWaterPointWithPath{}, err
			}

			path := []string{}
			if regionTitle.String == "" {
				path = []string{countryTitle, wwPoint.River.Title, wwPoint.Title}
			} else {
				path = []string{countryTitle, regionTitle.String, wwPoint.River.Title, wwPoint.Title}
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

func scanWwPointFull(rows *sql.Rows, additionalVars ...interface{}) (WhiteWaterPointFull, error) {
	wwp := WhiteWaterPointFull{}

	pointString := ""
	categoryString := ""
	lwCategoryString := ""
	mwCategoryString := ""
	hwCategoryString := ""
	lastAutoOrdering := pq.NullTime{}

	fields := append([]interface{}{&wwp.Id, &wwp.Title, &pointString, &categoryString, &wwp.ShortDesc, &wwp.Link,
		&wwp.River.Id, &wwp.River.Title,
		&lwCategoryString, &wwp.LowWaterDescription, &mwCategoryString, &wwp.MediumWaterDescription, &hwCategoryString, &wwp.HighWaterDescription,
		&wwp.Orient, &wwp.Approach, &wwp.Safety,
		&wwp.OrderIndex, &wwp.AutomaticOrdering, &lastAutoOrdering}, additionalVars...)

	err := rows.Scan(fields...)
	if err != nil {
		log.Errorf("Can not read from db: %v", err)
		return WhiteWaterPointFull{}, err
	}

	err = json.Unmarshal(categoryStrBytes(categoryString), &wwp.Category)
	if err != nil {
		log.Errorf("Can not parse category %s for white water object %d: %v", categoryString, wwp.Id, err)
		return WhiteWaterPointFull{}, err
	}

	err = json.Unmarshal(categoryStrBytes(lwCategoryString), &wwp.LowWaterCategory)
	if err != nil {
		log.Errorf("Can not parse low water category %s for white water object %d: %v", lwCategoryString, wwp.Id, err)
		return WhiteWaterPointFull{}, err
	}

	err = json.Unmarshal(categoryStrBytes(mwCategoryString), &wwp.MediumWaterCategory)
	if err != nil {
		log.Errorf("Can not parse medium water category %s for white water object %d: %v", mwCategoryString, wwp.Id, err)
		return WhiteWaterPointFull{}, err
	}

	err = json.Unmarshal(categoryStrBytes(hwCategoryString), &wwp.HighWaterCategory)
	if err != nil {
		log.Errorf("Can not parse high water category %s for white water object %d: %v", hwCategoryString, wwp.Id, err)
		return WhiteWaterPointFull{}, err
	}

	var pgPoint PgPoint
	err = json.Unmarshal([]byte(pointString), &pgPoint)
	if err != nil {
		log.Errorf("Can not parse point %s for white water object %d: %v", pointString, wwp.Id, err)
		return WhiteWaterPointFull{}, err
	}
	wwp.Point = pgPoint.Coordinates

	wwp.RiverId = wwp.River.Id

	if lastAutoOrdering.Valid {
		wwp.LastAutomaticOrdering = lastAutoOrdering.Time
	} else {
		wwp.LastAutomaticOrdering = util.ZeroDateUTC()
	}

	return wwp, err
}

func scanWwPoint(rows *sql.Rows, additionalVars ...interface{}) (WhiteWaterPoint, error) {
	var err error
	id := int64(-1)
	title := ""
	pointStr := ""
	categoryStr := ""
	shortDesc := sql.NullString{}
	link := sql.NullString{}
	riverId := sql.NullInt64{}

	fields := append([]interface{}{&id, &title, &pointStr, &categoryStr, &shortDesc, &link, &riverId}, additionalVars...)
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
	err = json.Unmarshal(categoryStrBytes(categoryStr), &category)
	if err != nil {
		log.Errorf("Can not parse category %s for white water object %d: %v", categoryStr, id, err)
		return WhiteWaterPoint{}, err
	}

	return WhiteWaterPoint{
		IdTitle: IdTitle{
			Id:id,
			Title: title,
		},
		RiverId:getOrElse(riverId, -1),
		Point: pgPoint.Coordinates,
		Category: category,
		ShortDesc: shortDesc.String,
		Link: link.String,
	}, nil
}

func categoryStrBytes(categoryStr string) []byte {
	if !(strings.HasPrefix(categoryStr, "\"") && strings.HasSuffix(categoryStr, "\"")) {
		categoryStr = "\"" + categoryStr + "\""
	}
	return []byte(categoryStr)
}

func (this whiteWaterStorage) InsertWhiteWaterPoints(whiteWaterPoints ...WhiteWaterPoint) error {
	return this.update(this.insertQuery, whiteWaterPoints...)
}

func (this whiteWaterStorage) UpdateWhiteWaterPoints(whiteWaterPoints ...WhiteWaterPoint) error {
	return this.update(this.updateQuery, whiteWaterPoints...)
}

func (this whiteWaterStorage) update(query string, whiteWaterPoints ...WhiteWaterPoint) error {
	vars := make([]interface{}, len(whiteWaterPoints))
	for i, p := range whiteWaterPoints {
		vars[i] = p
	}
	return this.performUpdates(query,
		func(entity interface{}) ([]interface{}, error) {
			wwp := entity.(WhiteWaterPoint)
			pathBytes, err := json.Marshal(geo.NewGeoPoint(wwp.Point))
			if err != nil {
				return nil, err
			}
			cat, err := wwp.Category.MarshalJSON()
			if err != nil {
				return nil, err
			}
			params := []interface{}{nullIf0(wwp.Id), wwp.Title, cat, string(pathBytes), wwp.ShortDesc, wwp.Link, nullIf0(wwp.RiverId)}
			return params, nil
		}, vars...)
}

func (this whiteWaterStorage) Remove(id int64, tx interface{}) error {
	log.Infof("Remove spot %d", id)
	return this.performUpdatesWithinTxOptionally(tx, this.deleteQuery, idMapper, id)
}

func (this whiteWaterStorage) RemoveByRiver(id int64, tx interface{}) error {
	log.Infof("Remove spots by river id", id)
	return this.performUpdatesWithinTxOptionally(tx, this.deleteForRiverQuery, idMapper, id)
}

func (this whiteWaterStorage) GetGeomCenterByRiver(riverId int64) (geo.Point, error) {
	p, found, err := this.doFindAndReturn(this.geomCenterByRiverQuery, func(rows *sql.Rows) (interface{}, error) {
		pointString := ""
		err := rows.Scan(&pointString)
		if err != nil {
			return geo.Point{}, err
		}
		var pgPoint PgPoint
		err = json.Unmarshal([]byte(pointString), &pgPoint)
		if err != nil {
			log.Errorf("Can not parse centroid point %s for river %d: %v", pointString, riverId, err)
			return geo.Point{}, err
		}
		return pgPoint.Coordinates, nil
	}, riverId)
	if err != nil {
		return geo.Point{}, err
	}
	if !found {
		return geo.Point{0, 0}, nil
	}
	return p.(geo.Point), nil
}

func (this whiteWaterStorage) Props() PropertyManager {
	return this.PropsManager
}

