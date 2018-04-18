package dao

import (
	"time"
	log "github.com/Sirupsen/logrus"
	"database/sql"
	_ "github.com/lib/pq"
	"encoding/json"
	"errors"
	"reflect"
	. "github.com/and-hom/wwmap/backend/geo"
	"fmt"
	"github.com/and-hom/wwmap/backend/model"
)

type Storage interface {
	FindRoute(id int64, route *Route) (bool, error)
	ListRoutes(bbox Bbox) []Route
	UpdateRoute(route Route) error
	AddRoute(route Route) (int64, error)
	DeleteRoute(id int64) error

	AddTracks(routeId int64, track ...Track) error
	UpdateTrack(track Track) error
	FindTrack(id int64, track *Track) (bool, error)
	FindTrackAsList(id int64) []Track
	FindTracksForRoute(routeId int64) []Track
	ListTracks(bbox Bbox) []Track
	DeleteTrack(id int64) error
	DeleteTracksForRoute(routeId int64) error

	AddEventPoint(routeId int64, eventPoint EventPoint) (int64, error)
	AddEventPoints(routeId int64, eventPoints ...EventPoint) error
	UpdateEventPoint(eventPoint EventPoint) error
	DeleteEventPoint(id int64) error
	DeleteEventPointsForRoute(routeId int64) error
	FindEventPoint(id int64, eventPoint *EventPoint) (bool, error)
	ListPoints(bbox Bbox) []EventPoint
	FindEventPointsForRoute(routeId int64) []EventPoint

	AddWaterWays(waterways ...WaterWay) error
	NearestWaterWays(point Point, limit int) ([]WaterWayTitle, error)

	// tmp
	AddTmpWaterWay(wwts ...WaterWayTmp) error
	AddTmpRef(pnts ...PointRef) error
	GetUniquePointRefIds() ([]int64, error)
	// end of tmp

	AddWhiteWaterPoints(whiteWaterPoint ...WhiteWaterPoint) error
	ListWhiteWaterPoints(bbox Bbox) []WhiteWaterPoint
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) Storage {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Can not connect to postgres: %v", err)
	}
	return PostgresStorage{
		db:db,
	}
}

func (s PostgresStorage) FindRoute(id int64, route *Route) (bool, error) {
	return s.doFind("SELECT id,title,COALESCE(category,'') FROM route WHERE id=$1", func(rows *sql.Rows) error {
		var categoryStr string
		var err error
		err = rows.Scan(&(route.Id), &(route.Title), &categoryStr)
		if err != nil {
			return err
		}
		err = route.Category.UnmarshalJSON([]byte(categoryStr))
		if err != nil {
			return err
		}
		return nil
	}, id)
}

func (s PostgresStorage) ListRoutes(bbox Bbox) []Route {
	return s.listRoutesInternal(`id in (SELECT route_id FROM track WHERE path && ST_MakeEnvelope($1,$2,$3,$4)
	UNION ALL SELECT route_id FROM point WHERE point && ST_MakeEnvelope($1,$2,$3,$4))`,
		bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s PostgresStorage) listRoutesInternal(whereClause string, queryParams ...interface{}) []Route {
	result, err := s.doFindList("SELECT id,title,COALESCE(category,'') FROM route WHERE " + whereClause,
		func(rows *sql.Rows) (Route, error) {
			var err error
			route := Route{}
			var categoryStr string
			err = rows.Scan(&(route.Id), &(route.Title), &categoryStr)
			if err != nil {
				return Route{}, err
			}
			err = route.Category.UnmarshalJSON([]byte(categoryStr))
			if err != nil {
				return Route{}, err
			}
			log.Errorf("%v", route)
			return route, nil
		}, queryParams...)
	if err != nil {
		log.Error("Can not load route list", err)
		return []Route{}
	}
	return result.([]Route)
}

func (s PostgresStorage) UpdateRoute(route Route) error {
	return s.performUpdates("UPDATE route SET title=$2, category=$3 WHERE id=$1",
		func(entity interface{}) ([]interface{}, error) {
			r := entity.(Route)
			catJson, err := r.Category.MarshalJSON()
			if err != nil {
				return nil, err
			}
			return []interface{}{r.Id, r.Title, catJson}, nil;
		}, route)
}

func (s PostgresStorage) AddRoute(route Route) (int64, error) {
	log.Info("Inserting route")
	categoryBytes, err := route.Category.MarshalJSON()
	if err != nil {
		return -1, err
	}
	return s.insertReturningId("INSERT INTO route(title,category) VALUES($1,$2) RETURNING id", route.Title, string(categoryBytes))
}

func (s PostgresStorage) DeleteRoute(id int64) error {
	log.Infof("Delete route %s", id)
	return s.performUpdates("DELETE FROM route WHERE id=$1", func(_id interface{}) ([]interface{}, error) {
		return []interface{}{_id}, nil;
	}, id)
}

func (s PostgresStorage) FindTrack(id int64, track *Track) (bool, error) {
	return s.doFind("SELECT id,type,title, length FROM track WHERE id=$1", func(rows *sql.Rows) error {
		var err error
		_type := ""
		rows.Scan(&(track.Id), &_type, &(track.Title), &(track.Length))

		track.Type, err = ParseTrackType(_type)
		return err
	}, id)
}

func (s PostgresStorage) FindTrackAsList(id int64) []Track {
	return s.listTracksInternal("id=$1", id)
}

func (s PostgresStorage) FindTracksForRoute(routeId int64) []Track {
	return s.listTracksInternal("route_id=$1 ORDER BY start_time ASC", routeId)
}

func (s PostgresStorage) ListTracks(bbox Bbox) []Track {
	return s.listTracksInternal("path && ST_MakeEnvelope($1,$2,$3,$4)", bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s PostgresStorage) listTracksInternal(whereClause string, queryParams ...interface{}) []Track {
	result, err := s.doFindList("SELECT id,type,title, ST_AsGeoJSON(path) as path, length FROM track WHERE " + whereClause,
		func(rows *sql.Rows) (Track, error) {
			var err error
			var _type string
			var pathStr string
			track := Track{}
			rows.Scan(&(track.Id), &_type, &(track.Title), &pathStr, &(track.Length))

			track.Type, err = ParseTrackType(_type)
			if err != nil {
				log.Error("Invalid track type", err)
				return Track{}, err
			}
			var path LineString
			err = json.Unmarshal([]byte(pathStr), &path)
			if err != nil {
				log.Errorf("Can not parse path for track %s: %v", path, err)
				return Track{}, err
			}
			track.Path = path.Coordinates
			return track, nil
		}, queryParams...)
	if err != nil {
		log.Error("Can not load track list", err)
		return []Track{}
	}
	return result.([]Track)
}

func (s PostgresStorage) AddTracks(routeId int64, tracks ...Track) error {
	log.Info("Inserting tracks")
	vars := make([]interface{}, len(tracks))
	for i, p := range tracks {
		vars[i] = p
	}
	return s.performUpdates(`INSERT INTO track(route_id, title, path, length, type, start_time, end_time)
	VALUES ($1 ,$2, ST_GeomFromGeoJSON($3), ST_Length(ST_GeomFromGeoJSON($3)::geography), $4, $5, $6)`,
		func(entity interface{}) ([]interface{}, error) {
			t := entity.(Track)

			pathBytes, err := json.Marshal(NewLineString(t.Path))
			if err != nil {
				return nil, err
			}
			return []interface{}{routeId, t.Title, string(pathBytes), string(t.Type),
				time.Time(t.StartTime), time.Time(t.EndTime)}, nil;
		}, vars...)
}

func (s PostgresStorage) UpdateTrack(track Track) error {
	log.Infof("Update track %d", track.Id)
	return s.performUpdates("UPDATE track SET title=$2, type=$3 WHERE id=$1",
		func(entity interface{}) ([]interface{}, error) {
			t := entity.(Track)
			return []interface{}{t.Id, t.Title, string(t.Type)}, nil;
		}, track)
}

func (s PostgresStorage) DeleteTrack(id int64) error {
	log.Infof("Delete track %s", id)
	return s.performUpdates("DELETE FROM track WHERE id=$1", func(_id interface{}) ([]interface{}, error) {
		return []interface{}{_id}, nil;
	}, id)
}

func (s PostgresStorage) DeleteTracksForRoute(routeId int64) error {
	log.Infof("Delete all tracks for route %s", routeId)
	return s.performUpdates("DELETE FROM track WHERE route_id=$1", func(_id interface{}) ([]interface{}, error) {
		return []interface{}{_id}, nil;
	}, routeId)
}

func (s PostgresStorage) AddEventPoints(routeId int64, eventPoints ...EventPoint) error {
	log.Info("Inserting eventPoints")
	vars := make([]interface{}, len(eventPoints))
	for i, p := range eventPoints {
		vars[i] = p
	}
	return s.performUpdates("INSERT INTO point(route_id, type, title, point, content, time) " +
		"VALUES ($1, $2, $3, ST_GeomFromGeoJSON($4), $5, $6)",
		func(entity interface{}) ([]interface{}, error) {
			p := entity.(EventPoint)

			pointBytes, err := json.Marshal(NewGeoPoint(p.Point))
			if err != nil {
				return nil, err
			}
			if err != nil {
				return nil, err
			}
			return []interface{}{routeId, string(p.Type), p.Title,
				string(pointBytes), p.Content, time.Time(p.Time)}, nil;
		}, vars...)
}

func (s PostgresStorage) AddEventPoint(routeId int64, eventPoint EventPoint) (int64, error) {
	log.Info("Inserting event point")
	pointBytes, err := json.Marshal(NewGeoPoint(eventPoint.Point))
	if err != nil {
		return -1, err
	}
	return s.insertReturningId("INSERT INTO point(route_id, type, title, point, content, time) " +
		"VALUES ($1, $2, $3, ST_GeomFromGeoJSON($4), $5, $6) RETURNING id", routeId, string(eventPoint.Type), eventPoint.Title,
		string(pointBytes), eventPoint.Content, time.Time(eventPoint.Time))
}

func (s PostgresStorage) UpdateEventPoint(eventPoint EventPoint) (error) {
	log.Info("Update event point")
	return s.performUpdates("UPDATE point SET type=$2, title=$3, point=ST_GeomFromGeoJSON($4), content=$5 WHERE id=$1",
		func(entity interface{}) ([]interface{}, error) {
			p := entity.(EventPoint)

			pointBytes, err := json.Marshal(NewGeoPoint(p.Point))
			if err != nil {
				return nil, err
			}

			return []interface{}{p.Id, string(p.Type), p.Title, string(pointBytes), p.Content}, nil
		}, eventPoint)
}

func (s PostgresStorage) DeleteEventPoint(id int64) error {
	log.Infof("Delete event point %s", id)
	return s.performUpdates("DELETE FROM point WHERE id=$1", func(_id interface{}) ([]interface{}, error) {
		return []interface{}{_id}, nil;
	}, id)
}

func (s PostgresStorage) DeleteEventPointsForRoute(routeId int64) error {
	log.Infof("Delete all event points for route %s", routeId)
	return s.performUpdates("DELETE FROM point WHERE route_id=$1", func(_id interface{}) ([]interface{}, error) {
		return []interface{}{_id}, nil;
	}, routeId)
}

func (s PostgresStorage)FindEventPoint(id int64, eventPoint *EventPoint) (bool, error) {
	return s.doFind(
		"SELECT id,type,title,content,ST_AsGeoJSON(point) as point,time FROM point WHERE id=$1",
		func(rows *sql.Rows) error {
			var err error

			var _type string
			var pointStr string
			rows.Scan(&(eventPoint.Id), &_type, &(eventPoint.Title),
				&(eventPoint.Content), &pointStr, &(eventPoint.Time))

			eventPoint.Type, err = ParseEventPointType(_type)
			if err != nil {
				return err
			}

			var pgPoint PgPoint
			err = json.Unmarshal([]byte(pointStr), &pgPoint)
			if err != nil {
				log.Errorf("Can not parse point for track %s: %v", pointStr, err)
				return err
			}
			eventPoint.Point = pgPoint.Coordinates
			return nil
		},
		id)
}

func (s PostgresStorage) ListPoints(bbox Bbox) []EventPoint {
	return s.listPointsInternal("point && ST_MakeEnvelope($1,$2,$3,$4)", bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s PostgresStorage) FindEventPointsForRoute(routeId int64) []EventPoint {
	return s.listPointsInternal("route_id=$1", routeId)
}

func (s PostgresStorage) listPointsInternal(whereClause string, queryParams ...interface{}) []EventPoint {
	result, err := s.doFindList("SELECT id, type, title, content, ST_AsGeoJSON(point) as point, time FROM point WHERE " +
		whereClause, func(rows *sql.Rows) (EventPoint, error) {
		var err error
		id := int64(-1)
		_type := ""
		title := ""
		content := ""
		t := time.Now() // any stub time
		pointStr := ""
		eventPoint := EventPoint{}
		rows.Scan(&id, &_type, &title, &content, &pointStr, &t)

		eventPoint.Id = id
		eventPoint.Type, err = ParseEventPointType(_type)
		if err != nil {
			log.Errorf("Can not parse point type %s for point %d: %v", _type, id, err)
			return EventPoint{}, err
		}
		eventPoint.Title = title
		eventPoint.Content = content

		var pgPoint PgPoint
		err = json.Unmarshal([]byte(pointStr), &pgPoint)
		if err != nil {
			log.Errorf("Can not parse point %s for point %d: %v", pointStr, id, err)
			return EventPoint{}, err
		}
		eventPoint.Point = pgPoint.Coordinates
		eventPoint.Time = JSONTime(t)
		return eventPoint, nil
	}, queryParams...)
	if (err != nil ) {
		return []EventPoint{}
	}
	return result.([]EventPoint)
}

func (this PostgresStorage) AddWaterWays(waterways ...WaterWay) error {
	vars := make([]interface{}, len(waterways))
	for i, p := range waterways {
		vars[i] = p
	}
	return this.performUpdates("INSERT INTO waterway(id, title, type, comment, path, verified, popularity) VALUES ($1, $2, $3, $4, ST_GeomFromGeoJSON($5), $6, $7)",
		func(entity interface{}) ([]interface{}, error) {
			waterway := entity.(WaterWay)

			pathBytes, err := json.Marshal(NewLineString(waterway.Path))
			if err != nil {
				return nil, err
			}
			return []interface{}{waterway.Id, waterway.Title, waterway.Type, waterway.Comment, string(pathBytes), waterway.Verified, waterway.Popularity}, nil;
		}, vars...)
}

func (this PostgresStorage) NearestWaterWays(point Point, limit int) ([]WaterWayTitle, error) {
	pointBytes, err := json.Marshal(NewGeoPoint(point))
	if err != nil {
		return []WaterWayTitle{}, err
	}
	result, err := this.doFindList("SELECT id, title FROM waterway ORDER BY ST_Distance(path,  ST_GeomFromGeoJSON($1)) LIMIT $2",
		func(rows *sql.Rows) (WaterWayTitle, error) {
			id := int64(-1)
			title := ""
			err := rows.Scan(&id, &title)
			if err != nil {
				return WaterWayTitle{}, err
			}

			return WaterWayTitle{
				Id:id,
				Title:title,
			}, nil
		}, string(pointBytes), limit)
	if (err != nil ) {
		return []WaterWayTitle{}, err
	}
	return result.([]WaterWayTitle), nil
}

func (this PostgresStorage) AddTmpWaterWay(wwts ...WaterWayTmp) error {
	vars := make([]interface{}, len(wwts))
	for i, p := range wwts {
		vars[i] = p
	}
	return this.performUpdates("INSERT INTO waterway_tmp(id, title, type, comment) VALUES ($1, $2, $3, $4)",
		func(entity interface{}) ([]interface{}, error) {
			waterway := entity.(WaterWayTmp)
			return []interface{}{waterway.Id, waterway.Title, waterway.Type, waterway.Comment}, nil;
		}, vars...)
}

func (this PostgresStorage) AddTmpRef(pnts ...PointRef) error {
	vars := make([]interface{}, len(pnts))
	for i, p := range pnts {
		vars[i] = p
	}
	return this.performUpdates("INSERT INTO point_ref_tmp(id, parent_id, idx) VALUES ($1,$2,$3)",
		func(entity interface{}) ([]interface{}, error) {
			pnt := entity.(PointRef)
			return []interface{}{pnt.Id, pnt.ParentId, pnt.Idx}, nil;
		}, vars...)
}

func (this PostgresStorage) GetUniquePointRefIds() ([]int64, error) {
	result, err := this.doFindList("SELECT distinct(id) FROM point_ref_tmp ",
		func(rows *sql.Rows) (int64, error) {
			id := int64(0)
			err := rows.Scan(&(id))
			if err != nil {
				return 0, err
			}
			return id, nil
		})
	if err != nil {
		log.Error("Can not load point ids list", err)
		return []int64{}, err
	}
	return result.([]int64), nil
}

func (this PostgresStorage) AddWhiteWaterPoints(whiteWaterPoints ...WhiteWaterPoint) error {
	vars := make([]interface{}, len(whiteWaterPoints))
	for i, p := range whiteWaterPoints {
		vars[i] = p
	}
	return this.performUpdates("INSERT INTO white_water_rapid(osm_id, title,type,category,comment,point,short_description, link, water_way_id) VALUES ($1, $2, $3, $4, $5, ST_GeomFromGeoJSON($6), $7, $8, $9)",
		func(entity interface{}) ([]interface{}, error) {
			wwp := entity.(WhiteWaterPoint)
			categoryBytes, err := wwp.Category.MarshalJSON()
			if err != nil {
				return nil, err
			}
			pathBytes, err := json.Marshal(NewGeoPoint(wwp.Point))
			if err != nil {
				return nil, err
			}
			fmt.Printf("id = %d", wwp.Id)
			return []interface{}{wwp.OsmId, wwp.Title, wwp.Type, string(categoryBytes), wwp.Comment, string(pathBytes), wwp.ShortDesc, wwp.Link, nullIf0(wwp.WaterWayId)}, nil
		}, vars...)
}

func nullIf0(x int64) sql.NullInt64 {
	if x==0 {
		return sql.NullInt64{
			Int64:0,
			Valid:false,
		}
	}
	return sql.NullInt64{
		Int64:x,
		Valid:true,
	}
}

func getOrElse(val sql.NullInt64, _default int64) int64 {
	if val.Valid {
		return val.Int64
	} else {
		return _default
	}
}

func (this PostgresStorage) ListWhiteWaterPoints(bbox Bbox) []WhiteWaterPoint {
	result, err := this.doFindList("SELECT id, osm_id, water_way_id, type, title, comment, ST_AsGeoJSON(point) as point, category, short_description, link FROM white_water_rapid WHERE point && ST_MakeEnvelope($1,$2,$3,$4)",
		func(rows *sql.Rows) (WhiteWaterPoint, error) {
			var err error
			id := int64(-1)
			osmId := sql.NullInt64{}
			waterWayId := sql.NullInt64{}
			_type := ""
			title := ""
			comment := ""
			pointStr := ""
			categoryStr := ""
			shortDesc := sql.NullString{}
			link := sql.NullString{}
			err = rows.Scan(&id, &osmId, &waterWayId, &_type, &title, &comment, &pointStr, &categoryStr, &shortDesc, &link)
			if err != nil {
				log.Errorf("Can not read from db: %v", err)
				return WhiteWaterPoint{}, err
			}

			var pgPoint PgPoint
			fmt.Printf("\n======================= %d %v %v %s %s %s %s %s\n\n", id, osmId, waterWayId, _type, title, comment, pointStr, categoryStr)
			err = json.Unmarshal([]byte(pointStr), &pgPoint)
			if err != nil {
				log.Errorf("Can not parse point %s for white water object %d: %v", pointStr, id, err)
				return WhiteWaterPoint{}, err
			}

			var category model.SportCategory
			err = json.Unmarshal([]byte(categoryStr), &category)
			if err != nil {
				log.Errorf("Can not parse category %s for white water object %d: %v", categoryStr, id, err)
				return WhiteWaterPoint{}, err
			}

			eventPoint := WhiteWaterPoint{
				Id:id,
				OsmId:getOrElse(osmId, -1),
				WaterWayId:getOrElse(waterWayId, -1),
				Title: title,
				Type: _type,
				Point: pgPoint.Coordinates,
				Comment: comment,
				Category: category,
				ShortDesc: shortDesc.String,
				Link: link.String,
			}
			return eventPoint, nil
		}, bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
	if (err != nil ) {
		return []WhiteWaterPoint{}
	}
	return result.([]WhiteWaterPoint)
}

func (s PostgresStorage)doFind(query string, callback func(rows *sql.Rows) error, args ...interface{}) (bool, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		err = callback(rows)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (s PostgresStorage)doFindList(query string, callback interface{}, args ...interface{}) (interface{}, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return []interface{}{}, err
	}
	defer rows.Close()

	funcValue := reflect.ValueOf(callback)
	returnType := funcValue.Type().Out(0)
	var result = reflect.MakeSlice(reflect.SliceOf(returnType), 0, 0)

	for rows.Next() {
		val := funcValue.Call([]reflect.Value{reflect.ValueOf(rows)})
		if val[1].Interface() == nil {
			result = reflect.Append(result, val[0])
		} else {
			log.Error(val[1])
		}
	}
	return result.Interface(), nil
}

func (s PostgresStorage)insertReturningId(query string, args ...interface{}) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return -1, err
	}
	rows, err := stmt.Query(args...)
	if err != nil {
		return -1, err
	}

	lastId := int64(-1)
	for rows.Next() {
		rows.Scan(&lastId)
	}

	err = rows.Close()
	if err != nil {
		return -1, err
	}
	err = stmt.Close()
	if err != nil {
		return -1, err
	}
	err = tx.Commit()
	if err != nil {
		return -1, err
	}
	if lastId < 0 {
		return -1, errors.New("Not inserted")
	}
	return lastId, nil

}

func (s PostgresStorage)performUpdates(query string, mapper func(entity interface{}) ([]interface{}, error), values ...interface{}) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	for _, entity := range values {
		values, err := mapper(entity)
		if err != nil {
			log.Errorf("Can not update %v", err)
			return err
		}
		_, err = stmt.Exec(values...)
		if err != nil {
			log.Errorf("Can not update %v", err)
			return err
		}
	}

	log.Infof("Update completed. Commit.")
	err = stmt.Close()
	if err != nil {
		return err
	}
	return tx.Commit()
}

type PgPoint struct {
	Coordinates Point `json:"coordinates"`
}
