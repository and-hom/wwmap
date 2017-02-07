package main

import (
	"time"
	log "github.com/Sirupsen/logrus"
	"database/sql"
	_ "github.com/lib/pq"
	"encoding/json"
	"errors"
	"fmt"
)

type TrackList []Track;

type Storage interface {
	FindTrackAsList(id int64) TrackList
	FindTracksForRoute(routeId int64) TrackList
	ListTracks(bbox Bbox) TrackList

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

func (s postgresStorage) FindTrackAsList(id int64) TrackList {
	return s.listTracksInternal("id=$1", id)
}

func (s postgresStorage) FindTracksForRoute(routeId int64) TrackList {
	return s.listTracksInternal("route_id=$1", routeId)
}

func (s postgresStorage) ListTracks(bbox Bbox) TrackList {
	return s.listTracksInternal("path && ST_MakeEnvelope($1,$2,$3,$4)", bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s postgresStorage) FindTrack(id int64, track *Track) (bool, error) {
	rows, err := s.db.Query("SELECT id,type,title FROM track WHERE id=$1", id)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		id := int64(-1)
		_type := ""
		title := ""
		rows.Scan(&id, &_type, &title)

		track.Id = id
		track.Type, err = parseTrackType(_type)
		if err != nil {
			return false, err
		}
		track.Title = title

		return true, nil
	}
	return false, nil
}

func (s postgresStorage) listTracksInternal(whereClause string, queryParams ...interface{}) TrackList {
	rows, err := s.db.Query("SELECT id,type,title, ST_AsGeoJSON(path) as path FROM track WHERE " +
		whereClause, queryParams...)
	if err != nil {
		log.Error("Can not load track list", err)
		return []Track{}
	}
	defer rows.Close()

	results := make([]Track, 0)
	for rows.Next() {
		id := int64(-1)
		_type := ""
		title := ""
		pathStr := ""
		rows.Scan(&id, &_type, &title, &pathStr)

		track := Track{}
		track.Id = id
		track.Type, err = parseTrackType(_type)
		if err != nil {
			log.Error("Invalid track type", err)
			return []Track{}
		}
		var path LineString
		err := json.Unmarshal([]byte(pathStr), &path)
		if err != nil {
			log.Errorf("Can not parse path for track %s: %v", path, err)
			continue
		}
		track.Path = path.Coordinates
		track.Title = title

		results = append(results, track)
	}
	return results
}

func (s postgresStorage) FindRoute(id int64, route *Route) (bool, error) {
	rows, err := s.db.Query("SELECT id,title FROM route WHERE id=$1", id)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		id := int64(-1)
		title := ""
		rows.Scan(&id, &title)

		route.Id = id
		route.Title = title

		return true, nil
	}
	return false, nil
}

func (s postgresStorage) ListRoutes(bbox Bbox) []Route {
	return s.listRoutesInternal(`id in (SELECT route_id FROM track WHERE path && ST_MakeEnvelope($1,$2,$3,$4)
	UNION ALL SELECT route_id FROM point WHERE point && ST_MakeEnvelope($1,$2,$3,$4))`,
		bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s postgresStorage) listRoutesInternal(whereClause string, queryParams ...interface{}) []Route {
	rows, err := s.db.Query("SELECT id,title FROM route WHERE " +
		whereClause, queryParams...)
	if err != nil {
		log.Error("Can not load route list", err)
		return []Route{}
	}
	defer rows.Close()

	results := make([]Route, 0)
	for rows.Next() {
		id := int64(-1)
		title := ""
		rows.Scan(&id, &title)

		route := Route{}
		route.Id = id
		route.Title = title

		results = append(results, route)
	}
	return results
}

func (s postgresStorage) AddTracks(routeId int64, tracks ...Track) error {
	log.Info("Inserting tracks")
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO track(route_id, title, path, type) VALUES ($1 ,$2, ST_GeomFromGeoJSON($3), $4)")
	if err != nil {
		return err
	}

	for _, track := range tracks {
		pathBytes, err := json.Marshal(LineString{
			Coordinates:track.Path,
			Type:LINE_STRING,
		})
		if err != nil {
			return err
		}
		r, err := stmt.Exec(routeId, track.Title, string(pathBytes), string(track.Type))
		if err != nil {
			return err
		}
		log.Info("Result is %v", r)
	}

	err = stmt.Close()
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s postgresStorage) UpdateRoute(route Route) error {
	log.Info("Update route")
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE route SET title=$2 WHERE id=$1")
	if err != nil {
		return err
	}

	result, err := stmt.Exec(route.Id, route.Title)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("Updated %d rows for route %d", rowsAffected, route.Id)
	}

	err = stmt.Close()
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s postgresStorage) AddRoute(route Route) (int64, error) {
	log.Info("Inserting event point")
	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}

	stmt, err := tx.Prepare("INSERT INTO route(title) VALUES($1) RETURNING id")
	if err != nil {
		return -1, err
	}

	rows, err := stmt.Query(route.Title)
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

func (s postgresStorage) UpdateTrack(track Track) error {
	log.Info("Update track")
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE track SET title=$2, type=$3 WHERE id=$1")
	if err != nil {
		return err
	}

	result, err := stmt.Exec(track.Id, track.Title, string(track.Type))
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("Updated %d rows for track %d", rowsAffected, track.Id)
	}

	err = stmt.Close()
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s postgresStorage) AddEventPoint(trackId int64, eventPoint EventPoint) (int64, error) {
	log.Info("Inserting event point")
	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}

	stmt, err := tx.Prepare("INSERT INTO point(track_id, type, title, point, content, time) " +
		"VALUES ($1, $2, $3, ST_GeomFromGeoJSON($4), $5, $6) RETURNING id")
	if err != nil {
		return -1, err
	}

	pointBytes, err := json.Marshal(NewGeoPoint(eventPoint.Point))
	if err != nil {
		return -1, err
	}

	rows, err := stmt.Query(trackId, string(eventPoint.Type), eventPoint.Title,
		string(pointBytes), eventPoint.Content, time.Time(eventPoint.Time))
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

func (s postgresStorage) UpdateEventPoint(eventPoint EventPoint) (error) {
	log.Info("Update event point")
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE point SET type=$2, title=$3, point=ST_GeomFromGeoJSON($4), content=$5 WHERE id=$1")
	if err != nil {
		return err
	}

	pointBytes, err := json.Marshal(NewGeoPoint(eventPoint.Point))
	if err != nil {
		return err
	}

	result, err := stmt.Exec(eventPoint.Id, string(eventPoint.Type), eventPoint.Title,
		string(pointBytes), eventPoint.Content)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("Updated %d rows for event point %d", rowsAffected, eventPoint.Id)
	}

	err = stmt.Close()
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (s postgresStorage) DeleteEventPoint(id int64) error {
	log.Infof("Delete event point %s", id)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM point WHERE id=$1", id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s postgresStorage)FindEventPoint(id int64, eventPoint *EventPoint) (bool, error) {
	rows, err := s.db.Query("SELECT id,type,title,content,ST_AsGeoJSON(point) as point,time" +
		" FROM point WHERE id=$1", id)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		id := int64(-1)
		_type := ""
		title := ""
		content := ""
		t := time.Now() // any stub time
		pointStr := ""
		rows.Scan(&id, &_type, &title, &content, &pointStr, &t)

		eventPoint.Id = id
		eventPoint.Type, err = parseEventPointType(_type)
		if err != nil {
			return false, err
		}
		eventPoint.Title = title
		eventPoint.Content = content

		var pgPoint PgPoint
		err = json.Unmarshal([]byte(pointStr), &pgPoint)
		if err != nil {
			log.Errorf("Can not parse point for track %s: %v", pointStr, err)
			continue
		}
		eventPoint.Point = pgPoint.Coordinates
		eventPoint.Time = JSONTime(t)
		return true, nil
	}
	return false, nil
}

func (s postgresStorage) ListPoints(bbox Bbox) []EventPoint {
	return s.listPointsInternal("point && ST_MakeEnvelope($1,$2,$3,$4)", bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s postgresStorage) FindEventPointsForRoute(routeId int64) []EventPoint {
	return s.listPointsInternal("route_id=$1", routeId)
}

func (s postgresStorage) listPointsInternal(whereClause string, queryParams ...interface{}) []EventPoint {
	rows, err := s.db.Query("SELECT id, type, title, content, ST_AsGeoJSON(point) as point, time FROM point WHERE " +
		whereClause, queryParams...)
	if err != nil {
		log.Error("Can not load point list", err)
		return []EventPoint{}
	}
	defer rows.Close()

	results := make([]EventPoint, 0)
	for rows.Next() {
		id := int64(-1)
		_type := ""
		title := ""
		content := ""
		t := time.Now() // any stub time
		pointStr := ""
		rows.Scan(&id, &_type, &title, &content, &pointStr, &t)

		eventPoint := EventPoint{}
		eventPoint.Id = id
		eventPoint.Type, err = parseEventPointType(_type)
		if err != nil {
			log.Errorf("Can not parse point type %s for point %d: %v", _type, id, err)
			continue
		}
		eventPoint.Title = title
		eventPoint.Content = content

		var pgPoint PgPoint
		err = json.Unmarshal([]byte(pointStr), &pgPoint)
		if err != nil {
			log.Errorf("Can not parse point %s for point %d: %v", pointStr, id, err)
			continue
		}
		eventPoint.Point = pgPoint.Coordinates
		eventPoint.Time = JSONTime(t)
		results = append(results, eventPoint)
	}
	return results
}

type PgPoint struct {
	Coordinates Point `json:"coordinates"`
}
