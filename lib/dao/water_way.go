package dao

import (
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func NewWaterWayPostgresDao(postgresStorage PostgresStorage) WaterWayDao {
	return waterWayStorage{
		PostgresStorage:                  postgresStorage,
		insertQuery:                      queries.SqlQuery("water-way", "insert"),
		updateQuery:                      queries.SqlQuery("water-way", "update"),
		listQuery:                        queries.SqlQuery("water-way", "list"),
		unlinkRiverQuery:                 queries.SqlQuery("water-way", "unlink-river"),
		detectForRiverQuery:              queries.SqlQuery("water-way", "detect-for-river"),
		bindToRiverQuery:                 queries.SqlQuery("water-way", "bind-to-river"),
		listByRiverIdsQuery:              queries.SqlQuery("water-way", "list-by-river-ids"),
		listWithRiverWithoutHeightsQuery: queries.SqlQuery("water-way", "list-with-river-without-heights"),
		listByRiverId4RouterQuery:        queries.SqlQuery("water-way", "list-by-river-id-4-router"),
		listByBbox4RouterQuery:           queries.SqlQuery("water-way", "list-by-bbox-4-router"),
		listByBboxQuery:                  queries.SqlQuery("water-way", "list-by-bbox"),
		listByBboxWithHeightsQuery:       queries.SqlQuery("water-way", "list-by-bbox-with-heights"),
		listByBbox4CorrectionQuery:       queries.SqlQuery("water-way", "list-4-correction"),
		updatePathSimplifiedQuery:        queries.SqlQuery("water-way", "update-path-simplified"),
		updatePathHeightAndDistQuery:     queries.SqlQuery("water-way", "update-path-height-and-dists"),
		listRefPoints:                    queries.SqlQuery("water-way", "get-ref-points"),
	}
}

type waterWayStorage struct {
	PostgresStorage
	insertQuery                      string
	updateQuery                      string
	listQuery                        string
	unlinkRiverQuery                 string
	detectForRiverQuery              string
	bindToRiverQuery                 string
	listByRiverIdsQuery              string
	listWithRiverWithoutHeightsQuery string
	listByRiverId4RouterQuery        string
	listByBbox4RouterQuery           string
	listByBboxWithHeightsQuery       string
	listByBbox4CorrectionQuery       string
	listRefPoints                    string
	listByBboxQuery                  string
	updatePathSimplifiedQuery        string
	updatePathHeightAndDistQuery     string
}

func (this waterWayStorage) AddWaterWays(waterways ...WaterWay) error {
	vars := make([]interface{}, len(waterways))
	for i, p := range waterways {
		vars[i] = p
	}
	return this.PerformUpdates(this.insertQuery,
		func(entity interface{}) ([]interface{}, error) {
			waterway := entity.(WaterWay)

			pathBytes, err := json.Marshal(geo.NewPgLineString(waterway.Path))
			if err != nil {
				return nil, err
			}
			return []interface{}{waterway.OsmId, waterway.Title, waterway.Type, waterway.Comment, string(pathBytes)}, nil
		}, vars...)
}

func (this waterWayStorage) UpdateWaterWay(waterway WaterWay) error {
	return this.PerformUpdates("",
		func(entity interface{}) ([]interface{}, error) {
			waterway := entity.(WaterWay)

			pathBytes, err := json.Marshal(geo.NewPgLineString(waterway.Path))
			if err != nil {
				return nil, err
			}
			return []interface{}{string(pathBytes), waterway.OsmId}, nil
		}, waterway)
}

func (this waterWayStorage) ForEachWaterWay(transformer func(WaterWay) (WaterWay, error), tmpTable string) error {
	rows, err := this.db.Query("")
	if err != nil {
		return err
	}
	defer rows.Close()

	tx, err := this.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO " + tmpTable + "(id, osm_id, river_Id, title, type, comment, path) VALUES ($1, $2, $3, $4, $5, $6, ST_GeomFromGeoJSON($7))")
	if err != nil {
		return err
	}
	defer stmt.Close()

	i := 0
	for rows.Next() {
		waterWay, err := scanWaterWay(rows)
		if err != nil {
			return err
		}
		waterWayNew, err := transformer(waterWay)
		if err != nil {
			log.Errorf("Can not transofrm waterway %d: %v", waterWay.Id, err)
			return err
		}

		pathBytesNew, err := json.Marshal(geo.NewPgLineString(waterWayNew.Path))
		if err != nil {
			log.Errorf("Can not serialize path %v: %v", waterWayNew.Path, err)
			return err
		}
		riverIdNew := sql.NullInt64{
			Valid: waterWayNew.RiverId > 0,
			Int64: waterWayNew.RiverId,
		}
		stmt.Exec(waterWayNew.Id, waterWayNew.OsmId, riverIdNew, waterWayNew.Title, waterWayNew.Type, waterWayNew.Comment, string(pathBytesNew))
		i++
		if i%1000 == 0 {
			//	err := tx.Commit()
			//	if err != nil {
			//		log.Errorf("Can not commit: %v", err)
			//		return err
			//	}
			log.Info("Progress: ", i)
		}
	}
	tx.Commit()
	return nil
}

func scanWaterWay(rows *sql.Rows) (WaterWay, error) {
	waterWay := WaterWay{}
	osmId := sql.NullInt64{}
	riverId := sql.NullInt64{}
	pathStr := ""
	err := rows.Scan(&waterWay.Id, &osmId, &riverId, &waterWay.Title, &waterWay.Type, &waterWay.Comment, &pathStr)
	if err != nil {
		return WaterWay{}, err
	}
	if osmId.Valid {
		waterWay.OsmId = osmId.Int64
	}
	if riverId.Valid {
		waterWay.RiverId = riverId.Int64
	}
	var path geo.LineString
	err = json.Unmarshal([]byte(pathStr), &path)
	if err != nil {
		log.Errorf("Can not parse path \"%s\": %v", pathStr, err)
		return WaterWay{}, err
	}
	waterWay.Path = path.GetFlippedPath()

	return waterWay, nil
}

func scanWaterWayTiny(rows *sql.Rows) (WaterWay4Router, error) {
	waterWay := WaterWay4Router{}
	pathStr := ""
	refsStr := ""
	err := rows.Scan(&waterWay.Id, &pathStr, &refsStr)
	if err != nil {
		return WaterWay4Router{}, err
	}
	var path geo.LineString
	err = json.Unmarshal([]byte(pathStr), &path)
	if err != nil {
		log.Errorf("Can not parse path \"%s\": %v", pathStr, err)
		return WaterWay4Router{}, err
	}
	waterWay.Path = path.GetFlippedPath()

	return waterWay, nil
}

func scanWaterWayWithHeightNonFlipped(rows *sql.Rows) (WaterWayWithHeight, error) {
	waterWay := WaterWayWithHeight{}
	pathStr := ""
	segmentLength := 0.0
	var heights pq.Int64Array
	err := rows.Scan(&waterWay.Id, &pathStr, &heights, &segmentLength)
	if err != nil {
		return WaterWayWithHeight{}, err
	}
	var path geo.LineString
	err = json.Unmarshal([]byte(pathStr), &path)
	if err != nil {
		log.Errorf("Can not parse path \"%s\": %v", pathStr, err)
		return WaterWayWithHeight{}, err
	}
	waterWay.Path = path.Coordinates
	waterWay.Length = int(segmentLength / 1000)
	waterWay.Heights = heights

	return waterWay, nil
}

func scanWaterWay4RouterNonFlipped(rows *sql.Rows) (WaterWay4Router, error) {
	waterWay := WaterWay4Router{}
	pathStr := ""
	refsStr := ""
	err := rows.Scan(&waterWay.Id, &pathStr, &refsStr)
	if err != nil {
		return WaterWay4Router{}, err
	}
	var path geo.LineString
	err = json.Unmarshal([]byte(pathStr), &path)
	if err != nil {
		log.Errorf("Can not parse path \"%s\": %v", pathStr, err)
		return WaterWay4Router{}, err
	}
	var refs []PgWaterWayRef
	err = json.Unmarshal([]byte(refsStr), &refs)
	if err != nil {
		log.Errorf("Can not parse int64 array \"%s\": %v", refsStr, err)
		return WaterWay4Router{}, err
	}
	waterWay.Refs = make(map[int64][]geo.Point)
	for i := 0; i < len(refs); i++ {
		waterWay.Refs[refs[i].RefId] = append(waterWay.Refs[refs[i].RefId], refs[i].CrossPoint.Coordinates)
	}
	waterWay.Path = path.Coordinates
	waterWay.Bounds = path.GetBounds(4.0)
	return waterWay, nil
}

func (this waterWayStorage) UnlinkRiver(id int64, tx interface{}) error {
	return this.PerformUpdatesWithinTxOptionally(tx, this.unlinkRiverQuery, IdMapper, id)
}

const DETECTION_MIN_DISTANCE_METERS = 30

func (this waterWayStorage) DetectForRiver(riverId int64) ([]WaterWay, error) {
	log.Debug(this.detectForRiverQuery)
	result, err := this.DoFindList(this.detectForRiverQuery, scanWaterWay, riverId, DETECTION_MIN_DISTANCE_METERS)
	if err != nil {
		return []WaterWay{}, err
	}
	return result.([]WaterWay), err
}

func (this waterWayStorage) BindWaterwaysToRivers() error {
	return this.PerformUpdates(this.bindToRiverQuery, IdMapper, 300)
}

func (this waterWayStorage) ListByRiverIds(riverIds ...int64) ([]WaterWay, error) {
	result, err := this.DoFindList(this.listByRiverIdsQuery, scanWaterWay, pq.Array(riverIds))
	if err != nil {
		return []WaterWay{}, err
	}
	return result.([]WaterWay), err
}

func (this waterWayStorage) ListWithRiver() ([]WaterWay4Router, error) {
	result, err := this.DoFindList(this.listWithRiverWithoutHeightsQuery, scanWaterWayTiny)
	if err != nil {
		return []WaterWay4Router{}, err
	}
	return result.([]WaterWay4Router), err
}

func (this waterWayStorage) ListWithHeightsByBbox(bbox geo.Bbox) ([]WaterWayWithHeight, error) {
	result, err := this.DoFindList(
		this.listByBboxWithHeightsQuery,
		scanWaterWayWithHeightNonFlipped,
		bbox.Y1,
		bbox.X1,
		bbox.Y2,
		bbox.X2,
	)
	if err != nil {
		return []WaterWayWithHeight{}, err
	}
	return result.([]WaterWayWithHeight), err
}

func (this waterWayStorage) ListByBboxNonFilpped(bbox geo.Bbox) ([]WaterWay4Router, error) {
	result, err := this.DoFindList(this.listByBbox4RouterQuery, scanWaterWay4RouterNonFlipped, bbox.Y1, bbox.X1, bbox.Y2, bbox.X2)
	if err != nil {
		return []WaterWay4Router{}, err
	}
	return result.([]WaterWay4Router), err
}

func (this waterWayStorage) ListByRiverIdNonFlipped(riverId int64) ([]WaterWay4Router, error) {
	result, err := this.DoFindList(this.listByRiverId4RouterQuery, scanWaterWayTiny, riverId)
	if err != nil {
		return []WaterWay4Router{}, err
	}
	return result.([]WaterWay4Router), err
}

func (this waterWayStorage) ListByBbox(bbox geo.Bbox) ([]WaterWay, error) {
	result, err := this.DoFindList(this.listByBboxQuery, scanWaterWay, bbox.Y1, bbox.X1, bbox.Y2, bbox.X2)
	if err != nil {
		return []WaterWay{}, err
	}
	return result.([]WaterWay), err
}

type PgWaterWayRef struct {
	RefId      int64        `json:"id"`
	CrossPoint geo.GeoPoint `json:"cross_point"`
}

func (this waterWayStorage) List(limit int, offset int) ([]WaterWay4PathCorrection, error) {
	lst, err := this.DoFindList(this.listByBbox4CorrectionQuery, func(rows *sql.Rows) (WaterWay4PathCorrection, error) {
		waterway := WaterWay4PathCorrection{}
		pathString := ""
		pathSimplifiedString := ""
		err := rows.Scan(&waterway.Id, &pathString, &pathSimplifiedString)
		if err != nil {
			return waterway, err
		}

		if err := this.parsePath(waterway.Id, pathString, &waterway.Path); err != nil {
			return waterway, err
		}
		if err := this.parsePath(waterway.Id, pathSimplifiedString, &waterway.PathSimplified); err != nil {
			return waterway, err
		}

		return waterway, nil
	}, limit, offset)
	if err != nil {
		return []WaterWay4PathCorrection{}, err
	}

	waterways := lst.([]WaterWay4PathCorrection)
	ids := make([]int64, len(waterways))
	for i := 0; i < len(waterways); i++ {
		ids[i] = waterways[i].Id
	}

	pointsByRiverId := make(map[int64][]geo.Point)
	_, err = this.DoFindList(this.listRefPoints, func(rows *sql.Rows) (int64, error) {
		id := int64(0)
		pointStr := ""
		var p geo.PgPointOrLineString

		rows.Scan(&id, &pointStr)
		err := json.Unmarshal([]byte(pointStr), &p)
		if err != nil {
			log.Errorf("Can not parse ref point for waterway %d: %v", id, err)
			return 0, err
		}
		if p.Coordinates.Point == nil {
			log.Errorf("Ref point for waterway %d is not Point", id)
			return 0, errors.New("Is not a Point")
		}
		pointsByRiverId[id] = append(pointsByRiverId[id], *(p.Coordinates.Point))

		return id, nil
	}, pq.Array(ids))
	if err != nil {
		return []WaterWay4PathCorrection{}, err
	}

	for i := 0; i < len(waterways); i++ {
		points, found := pointsByRiverId[waterways[i].Id]
		if found {
			waterways[i].CrossPoints = points
		} else {
			waterways[i].CrossPoints = []geo.Point{}
		}
	}
	return waterways, nil
}

func (this waterWayStorage) parsePath(id int64, pathString string, target *[]geo.Point) error {
	var path geo.PgPointOrLineString
	err := json.Unmarshal([]byte(pathString), &path)
	if err != nil {
		log.Errorf("Can not parse path for waterway %d: %v", id, err)
		return err
	}
	if path.Coordinates.Line == nil {
		log.Errorf("Path for waterway %d is not LineString", id)
		return errors.New("Is not a LineString")
	}
	*target = *path.Coordinates.Line
	return nil
}

func (this waterWayStorage) PathSimplifiedPersister() (PathSimplifiedPersister, error) {
	stmt, err := this.db.Prepare(this.updatePathSimplifiedQuery)
	if err != nil {
		return PathSimplifiedPersisterImpl{}, err
	}
	return PathSimplifiedPersisterImpl{
		stmt: stmt,
	}, nil
}

type PathSimplifiedPersisterImpl struct {
	stmt *sql.Stmt
}

func (this PathSimplifiedPersisterImpl) Add(id int64, pathSimplified []geo.Point) error {
	pgPath := geo.LineString{
		Coordinates: pathSimplified,
		Type:        geo.LINE_STRING,
	}
	jsonBytes, err := json.Marshal(pgPath)
	if err != nil {
		return err
	}
	_, err = this.stmt.Exec(id, string(jsonBytes))
	return err
}

func (this PathSimplifiedPersisterImpl) Close() error {
	return this.stmt.Close()
}
