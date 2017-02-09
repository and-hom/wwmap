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
)

type Storage interface {
	FindTrackAsList(id int64) []Track
	FindTracksForRoute(routeId int64) []Track
	ListTracks(bbox Bbox) []Track

	FindRoute(id int64, route *Route) (bool, error)
	ListRoutes(bbox Bbox) []Route
	UpdateRoute(route Route) error
	AddRoute(route Route) (int64, error)

	AddTracks(routeId int64, track ...Track) error
	UpdateTrack(track Track) error
	FindTrack(id int64, track *Track) (bool, error)

	AddEventPoint(trackId int64, eventPoint EventPoint) (int64, error)
	UpdateEventPoint(eventPoint EventPoint) error
	DeleteEventPoint(id int64) error
	FindEventPoint(id int64, eventPoint *EventPoint) (bool, error)
	ListPoints(bbox Bbox) []EventPoint
	FindEventPointsForRoute(routeId int64) []EventPoint
}

type postgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() Storage {
	db, err := sql.Open("postgres", "postgres://wwmap:wwmap@localhost:5432/wwmap?sslmode=require")
	if err != nil {
		log.Fatalf("Can not connect to postgres: %v", err)
	}
	return postgresStorage{
		db:db,
	}
}

func (s postgresStorage) FindRoute(id int64, route *Route) (bool, error) {
	return s.doFind("SELECT id,title FROM route WHERE id=$1", func(rows *sql.Rows) error {
		rows.Scan(&(route.Id), &(route.Title))
		return nil
	}, id)
}

func (s postgresStorage) ListRoutes(bbox Bbox) []Route {
	return s.listRoutesInternal(`id in (SELECT route_id FROM track WHERE path && ST_MakeEnvelope($1,$2,$3,$4)
	UNION ALL SELECT route_id FROM point WHERE point && ST_MakeEnvelope($1,$2,$3,$4))`,
		bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s postgresStorage) listRoutesInternal(whereClause string, queryParams ...interface{}) []Route {
	result, err := s.doFindList("SELECT id,title FROM route WHERE " + whereClause,
		func(rows *sql.Rows) (Route, error) {
			route := Route{}
			rows.Scan(&(route.Id), &(route.Title))
			return route, nil
		}, queryParams...)
	if err != nil {
		log.Error("Can not load route list", err)
		return []Route{}
	}
	return result.([]Route)
}

func (s postgresStorage) UpdateRoute(route Route) error {
	return s.performUpdates("UPDATE route SET title=$2 WHERE id=$1",
		func(entity interface{}) ([]interface{}, error) {
			r := entity.(Route)
			return []interface{}{r.Id, r.Title}, nil;
		}, route)
}

func (s postgresStorage) AddRoute(route Route) (int64, error) {
	log.Info("Inserting route")
	return s.insertReturningId("INSERT INTO route(title) VALUES($1) RETURNING id", route.Title)
}

func (s postgresStorage) FindTrack(id int64, track *Track) (bool, error) {
	return s.doFind("SELECT id,type,title FROM track WHERE id=$1", func(rows *sql.Rows) error {
		var err error
		_type := ""
		rows.Scan(&(track.Id), &_type, &(track.Title))

		track.Type, err = ParseTrackType(_type)
		return err
	}, id)
}

func (s postgresStorage) FindTrackAsList(id int64) []Track {
	return s.listTracksInternal("id=$1", id)
}

func (s postgresStorage) FindTracksForRoute(routeId int64) []Track {
	return s.listTracksInternal("route_id=$1", routeId)
}

func (s postgresStorage) ListTracks(bbox Bbox) []Track {
	return s.listTracksInternal("path && ST_MakeEnvelope($1,$2,$3,$4)", bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s postgresStorage) listTracksInternal(whereClause string, queryParams ...interface{}) []Track {
	result, err := s.doFindList("SELECT id,type,title, ST_AsGeoJSON(path) as path FROM track WHERE " + whereClause,
		func(rows *sql.Rows) (Track, error) {
			var err error
			var _type string
			var pathStr string
			track := Track{}
			rows.Scan(&(track.Id), &_type, &(track.Title), &pathStr)

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

func (s postgresStorage) AddTracks(routeId int64, tracks ...Track) error {
	log.Info("Inserting tracks")
	return s.performUpdates("INSERT INTO track(route_id, title, path, type) VALUES ($1 ,$2, ST_GeomFromGeoJSON($3), $4)", func(entity interface{}) ([]interface{}, error) {
		t := entity.(Track)

		pathBytes, err := json.Marshal(LineString{
			Coordinates:t.Path,
			Type:LINE_STRING,
		})
		if err != nil {
			return nil, err
		}
		return []interface{}{t.Id, t.Title, string(pathBytes), string(t.Type)}, nil;
	}, tracks)
}

func (s postgresStorage) UpdateTrack(track Track) error {
	log.Infof("Update track %d", track.Id)
	return s.performUpdates("UPDATE track SET title=$2, type=$3 WHERE id=$1",
		func(entity interface{}) ([]interface{}, error) {
			t := entity.(Track)
			return []interface{}{t.Id, t.Title, string(t.Type)}, nil;
		}, track)
}

func (s postgresStorage) AddEventPoint(routeId int64, eventPoint EventPoint) (int64, error) {
	log.Info("Inserting event point")
	pointBytes, err := json.Marshal(NewGeoPoint(eventPoint.Point))
	if err != nil {
		return -1, err
	}
	return s.insertReturningId("INSERT INTO point(route_id, type, title, point, content, time) " +
		"VALUES ($1, $2, $3, ST_GeomFromGeoJSON($4), $5, $6) RETURNING id", routeId, string(eventPoint.Type), eventPoint.Title,
		string(pointBytes), eventPoint.Content, time.Time(eventPoint.Time))
}

func (s postgresStorage) UpdateEventPoint(eventPoint EventPoint) (error) {
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

func (s postgresStorage) DeleteEventPoint(id int64) error {
	log.Infof("Delete event point %s", id)
	ch := make(chan []interface{})
	go func() {
		ch <- []interface{}{id}
		close(ch)
	}()
	return s.performUpdates("DELETE FROM point WHERE id=$1", func(_id interface{}) ([]interface{}, error) {
		return []interface{}{_id}, nil;
	}, id)
}

func (s postgresStorage)FindEventPoint(id int64, eventPoint *EventPoint) (bool, error) {
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

func (s postgresStorage) ListPoints(bbox Bbox) []EventPoint {
	return s.listPointsInternal("point && ST_MakeEnvelope($1,$2,$3,$4)", bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s postgresStorage) FindEventPointsForRoute(routeId int64) []EventPoint {
	return s.listPointsInternal("route_id=$1", routeId)
}

func (s postgresStorage) listPointsInternal(whereClause string, queryParams ...interface{}) []EventPoint {
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

func (s postgresStorage)doFind(query string, callback func(rows *sql.Rows) error, args ...interface{}) (bool, error) {
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

func (s postgresStorage)doFindList(query string, callback interface{}, args ...interface{}) (interface{}, error) {
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
		//obj, err := callback(rows)
		if err == nil {
			result = reflect.Append(result, val[0])
		}
	}
	return result.Interface(), nil
}

func (s postgresStorage)insertReturningId(query string, args ...interface{}) (int64, error) {
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

func (s postgresStorage)performUpdates(query string, mapper func(entity interface{}) ([]interface{}, error), values ...interface{}) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	for _, entity := range values {
		log.Infof("Update %v", entity)
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