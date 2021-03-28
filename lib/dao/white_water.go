package dao

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/model"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/lib/pq"
	"strings"
	"time"
)

func NewWhiteWaterPostgresDao(postgresStorage PostgresStorage) WhiteWaterDao {
	return whiteWaterStorage{
		PostgresStorage:            postgresStorage,
		PropsManager:               PropertyManagerImpl{table: queries.SqlQuery("white-water", "table"), dao: &postgresStorage},
		listByBoxQuery:             queries.SqlQuery("white-water", "by-box"),
		listByRiverQuery:           queries.SqlQuery("white-water", "by-river"),
		listByRiverFullQuery:       queries.SqlQuery("white-water", "by-river-full"),
		listByRiverAndTitleQuery:   queries.SqlQuery("white-water", "by-river-and-title"),
		insertQuery:                queries.SqlQuery("white-water", "insert"),
		insertFullQuery:            queries.SqlQuery("white-water", "insert-full"),
		updateQuery:                queries.SqlQuery("white-water", "update"),
		byIdQuery:                  queries.SqlQuery("white-water", "by-id"),
		byIdFullQuery:              queries.SqlQuery("white-water", "by-id-full"),
		updateFullQuery:            queries.SqlQuery("white-water", "update-full"),
		deleteQuery:                queries.SqlQuery("white-water", "delete"),
		deleteForRiverQuery:        queries.SqlQuery("white-water", "delete-for-river"),
		geomCenterByRiverQuery:     queries.SqlQuery("white-water", "geom-center-by-river"),
		riverBoundsQuery:           queries.SqlQuery("white-water", "river-bounds"),
		autoOrderingRiverIdsQuery:  queries.SqlQuery("white-water", "auto-ordering-river-ids"),
		distanceFromBeginningQuery: queries.SqlQuery("white-water", "distance-from-beginning"),
		updateOrderIdxQuery:        queries.SqlQuery("white-water", "update-order-idx"),
		findByTitlePartQuery:       queries.SqlQuery("white-water", "by-title-part"),
		parentIds:                  queries.SqlQuery("white-water", "parent-ids"),
	}
}

type whiteWaterStorage struct {
	PostgresStorage
	PropsManager               PropertyManager
	listByBoxQuery             string
	listByRiverQuery           string
	listByRiverFullQuery       string
	listByRiverAndTitleQuery   string
	findByTitlePartQuery       string
	insertQuery                string
	insertFullQuery            string
	updateQuery                string
	byIdQuery                  string
	byIdFullQuery              string
	updateFullQuery            string
	deleteQuery                string
	deleteForRiverQuery        string
	geomCenterByRiverQuery     string
	riverBoundsQuery           string
	autoOrderingRiverIdsQuery  string
	distanceFromBeginningQuery string
	updateOrderIdxQuery        string
	parentIds                  string
}

func (this whiteWaterStorage) ListByBbox(bbox geo.Bbox) ([]WhiteWaterPointWithRiverTitle, error) {
	return this.list(this.listByBoxQuery, bbox.Y1, bbox.X1, bbox.Y2, bbox.X2)
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

func (this whiteWaterStorage) FindByTitlePart(tPart string, limit, offset int, showUnpublished bool) ([]WhiteWaterPointWithRiverTitle, error) {
	tPart = eYoRepl.ReplaceAllLiteralString(tPart, "е")
	return this.list(this.findByTitlePartQuery, pq.Array(strings.Fields(tPart)), limit, offset, showUnpublished)
}

func (this whiteWaterStorage) Find(id int64) (WhiteWaterPointWithRiverTitle, bool, error) {
	found, err := this.list(this.byIdQuery, id)
	if err != nil {
		return WhiteWaterPointWithRiverTitle{}, false, err
	}
	if len(found) == 0 {
		return WhiteWaterPointWithRiverTitle{}, false, nil
	}
	return found[0], true, nil
}

func (this whiteWaterStorage) FindFull(id int64) (WhiteWaterPointFull, error) {
	result, found, err := this.DoFindAndReturn(this.byIdFullQuery, func(rows *sql.Rows) (interface{}, error) {
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

func (this whiteWaterStorage) InsertWhiteWaterPointFull(whiteWaterPoint WhiteWaterPointFull, tx interface{}) (int64, error) {
	ids, err := this.updateReturningColumnsWithinTxOptionally(tx, this.insertFullQuery, paramsFull, true, whiteWaterPoint)
	if err != nil {
		return 0, err
	}
	idsInt64 := this.getFirstColumnAsInt64(ids)
	if len(ids) == 0 {
		return 0, errors.New("No id of inserted row returned")
	}
	return idsInt64[0], nil
}

func (this whiteWaterStorage) UpdateWhiteWaterPointsFull(whiteWaterPoints ...WhiteWaterPointFull) error {
	vars := make([]interface{}, len(whiteWaterPoints))
	for i, p := range whiteWaterPoints {
		vars[i] = p
	}
	return this.PerformUpdates(this.updateFullQuery, WhiteWaterPointFullMapper, vars...)
}

func (this whiteWaterStorage) UpdateWhiteWaterPointFull(whiteWaterPoint WhiteWaterPointFull, tx interface{}) error {
	return this.PerformUpdatesWithinTxOptionally(tx, this.updateFullQuery, WhiteWaterPointFullMapper, whiteWaterPoint)
}

func WhiteWaterPointFullMapper(entity interface{}) ([]interface{}, error) {
	wwp := entity.(WhiteWaterPointFull)
	params, err := paramsFull(wwp)
	if err != nil {
		return nil, err
	}
	return append([]interface{}{wwp.Id}, params...), nil
}

func paramsFull(p interface{}) ([]interface{}, error) {
	wwp := p.(WhiteWaterPointFull)

	pointBytes, err := json.Marshal(wwp.Point.ToPg())
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
	aliasesB, err := json.Marshal(wwp.Aliases)
	if err != nil {
		return nil, err
	}
	propsB, err := json.Marshal(wwp.Props)
	if err != nil {
		return nil, err
	}

	return []interface{}{wwp.Title, string(cat), string(pointBytes), wwp.ShortDesc, wwp.Link, nullIf0(wwp.River.Id),
		string(lwCat), wwp.LowWaterDescription, string(mwCat), wwp.MediumWaterDescription, string(hwCat), wwp.HighWaterDescription,
		wwp.Orient, wwp.Approach, wwp.Safety, wwp.OrderIndex, wwp.AutomaticOrdering, string(aliasesB), string(propsB)}, nil
}

func (this whiteWaterStorage) list(query string, vars ...interface{}) ([]WhiteWaterPointWithRiverTitle, error) {
	result, err := this.DoFindList(query,
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
	if err != nil {
		return []WhiteWaterPointWithRiverTitle{}, err
	}
	return result.([]WhiteWaterPointWithRiverTitle), nil
}

func (this whiteWaterStorage) listFull(query string, vars ...interface{}) ([]WhiteWaterPointFull, error) {
	result, err := this.DoFindList(query, scanWwPointFull, vars...)
	if err != nil {
		return []WhiteWaterPointFull{}, err
	}
	return result.([]WhiteWaterPointFull), nil
}

func scanWwPointFull(rows *sql.Rows, additionalVars ...interface{}) (WhiteWaterPointFull, error) {
	wwp := WhiteWaterPointFull{}

	pointString := ""
	categoryString := ""
	lwCategoryString := ""
	mwCategoryString := ""
	hwCategoryString := ""
	lastAutoOrdering := pq.NullTime{}
	aliasesStr := ""
	props := ""

	fields := append([]interface{}{&wwp.Id, &wwp.Title, &pointString, &categoryString, &wwp.ShortDesc, &wwp.Link,
		&wwp.River.Id, &wwp.River.Title, &wwp.River.Region.Id, &wwp.River.Region.CountryId, &wwp.River.Region.Fake,
		&lwCategoryString, &wwp.LowWaterDescription, &mwCategoryString, &wwp.MediumWaterDescription, &hwCategoryString, &wwp.HighWaterDescription,
		&wwp.Orient, &wwp.Approach, &wwp.Safety,
		&wwp.OrderIndex, &wwp.AutomaticOrdering, &lastAutoOrdering, &aliasesStr, &props}, additionalVars...)

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

	var pgPoint geo.PgPointOrLineString
	err = json.Unmarshal([]byte(pointString), &pgPoint)
	if err != nil {
		log.Errorf("Can not parse point %s for white water object %d: %v", pointString, wwp.Id, err)
		return WhiteWaterPointFull{}, err
	}
	wwp.Point = pgPoint.Coordinates.Flip()

	wwp.RiverId = wwp.River.Id

	if lastAutoOrdering.Valid {
		wwp.LastAutomaticOrdering = lastAutoOrdering.Time
	} else {
		wwp.LastAutomaticOrdering = util.ZeroDateUTC()
	}

	err = json.Unmarshal([]byte(aliasesStr), &wwp.Aliases)
	if err != nil {
		log.Errorf("Can not parse aliases %s for white water object %d: %v", aliasesStr, wwp.Id, err)
		return WhiteWaterPointFull{}, err
	}

	err = json.Unmarshal([]byte(props), &wwp.Props)
	if err != nil {
		log.Errorf("Can not parse props %s for white water object %d: %v", props, wwp.Id, err)
		return WhiteWaterPointFull{}, err
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

	var pgPoint geo.PgPointOrLineString
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
			Id:    id,
			Title: title,
		},
		RiverId:   getOrElse(riverId, -1),
		Point:     pgPoint.Coordinates.Flip(),
		Category:  category,
		ShortDesc: shortDesc.String,
		Link:      link.String,
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
	return this.PerformUpdates(query,
		func(entity interface{}) ([]interface{}, error) {
			wwp := entity.(WhiteWaterPoint)
			pathBytes, err := json.Marshal(wwp.Point.ToPg())
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
	return this.PerformUpdatesWithinTxOptionally(tx, this.deleteQuery, IdMapper, id)
}

func (this whiteWaterStorage) RemoveByRiver(id int64, tx interface{}) error {
	log.Infof("Remove spots by river id", id)
	return this.PerformUpdatesWithinTxOptionally(tx, this.deleteForRiverQuery, IdMapper, id)
}

func (this whiteWaterStorage) GetGeomCenterByRiver(riverId int64) (*geo.Point, error) {
	p, found, err := this.DoFindAndReturn(this.geomCenterByRiverQuery, func(rows *sql.Rows) (interface{}, error) {
		pointString := ""
		err := rows.Scan(&pointString)
		if err != nil {
			return geo.Point{}, err
		}
		var pgPoint geo.GeoPoint
		err = json.Unmarshal([]byte(pointString), &pgPoint)
		if err != nil {
			log.Errorf("Can not parse centroid point %s for river %d: %v", pointString, riverId, err)
			return nil, err
		}
		coords := pgPoint.Coordinates.Flip()
		return &coords, nil
	}, riverId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return p.(*geo.Point), nil
}

func (this whiteWaterStorage) GetRiverBounds(riverId int64) (*geo.Bbox, error) {
	p, found, err := this.DoFindAndReturn(this.riverBoundsQuery, func(rows *sql.Rows) (interface{}, error) {
		pointString := ""
		err := rows.Scan(&pointString)
		if err != nil {
			return nil, err
		}
		bounds, err := ParseBounds(pointString)
		return &bounds, err
	}, riverId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return p.(*geo.Bbox), nil
}

func (this whiteWaterStorage) Props() PropertyManager {
	return this.PropsManager
}

func (this whiteWaterStorage) AutoOrderingRiverIds() ([]int64, error) {
	r, err := this.DoFindList(this.autoOrderingRiverIdsQuery, Int64ColumnMapper)
	if err != nil {
		return []int64{}, err
	}
	return r.([]int64), err
}

type idDistPair struct {
	int64
	int
}

const ORDERING_MIN_DISTANCE_METERS = 30

func (this whiteWaterStorage) DistanceFromBeginning(riverId int64, path []geo.Point) (map[int64]int, error) {
	result := make(map[int64]int)
	pathB, err := json.Marshal(geo.NewPgLineString(path))
	if err != nil {
		return result, err
	}
	pairs, err := this.DoFindList(this.distanceFromBeginningQuery, func(rows *sql.Rows) (idDistPair, error) {
		wwpId := int64(0)
		dist := 0
		err := rows.Scan(&wwpId, &dist)
		return idDistPair{wwpId, dist}, err
	}, riverId, string(pathB), ORDERING_MIN_DISTANCE_METERS)
	if err != nil {
		return result, err
	}
	for _, pair := range pairs.([]idDistPair) {
		result[pair.int64] = pair.int
	}
	return result, nil
}

func (this whiteWaterStorage) UpdateOrderIdx(idx map[int64]int) error {
	now := time.Now()
	params := make([]interface{}, 0, len(idx))
	for id, val := range idx {
		params = append(params, []interface{}{id, val, now})
	}
	return this.PerformUpdates(this.updateOrderIdxQuery, ArrayMapper, params...)
}

func (this whiteWaterStorage) GetParentIds(spotIds []int64) (map[int64]SpotParentIds, error) {
	result := make(map[int64]SpotParentIds)

	_, err := this.DoFindList(this.parentIds, func(rows *sql.Rows) (int, error) {
		spotId := int64(0)
		parentIds := SpotParentIds{}
		err := rows.Scan(&spotId, &parentIds.RiverId, &parentIds.RegionId, &parentIds.CountryId, &parentIds.SpotTitle, &parentIds.RiverTitle)
		if err == nil {
			result[spotId] = parentIds
		}
		return 0, err
	}, pq.Array(spotIds))

	if err != nil {
		return result, err
	}
	return result, nil
}
