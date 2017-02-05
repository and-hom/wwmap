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

func (t TrackList) withoutPath() TrackList {
	newList := make([]Track, len(t))
	for i := 0; i < len(t); i++ {
		newList[i] = Track{
			Id:t[i].Id,
			Title:t[i].Title,
			Points:t[i].Points,
		}
	}
	return newList
}

type Storage interface {
	getTrack(id int64) TrackList
	getTracks(bbox Bbox) TrackList

	AddTracks(track ...Track) error
	UpdateTrack(track Track) error
	FindTrack(id int64, track *Track) (bool, error) 

	AddEventPoint(trackId int64, eventPoint EventPoint) (int64, error)
	UpdateEventPoint(eventPoint EventPoint) error
	DeleteEventPoint(id int64) error
	FindEventPoint(id int64, eventPoint *EventPoint) (bool, error)
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

func (s postgresStorage) getTrack(id int64) TrackList {
	return s.getTracksInternal("t.id=$1", id)
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

func (s postgresStorage) getTracks(bbox Bbox) TrackList {
	return s.getTracksInternal("path && ST_MakeEnvelope($1,$2,$3,$4)", bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
}

func (s postgresStorage) getTracksInternal(whereClause string, queryParams ...interface{}) TrackList {
	rows, err := s.db.Query(`SELECT t.id, t.title, ST_AsGeoJSON(t.path) as path, t.type as type,
	COALESCE(p.id, -1), COALESCE(p.type,''), COALESCE(p.title,''), COALESCE(p.content,''),
	COALESCE(ST_AsGeoJSON(p.point),'') as point, COALESCE(p.time, now())
	FROM track t LEFT OUTER JOIN point p ON t.id=p.track_id
	WHERE ` + whereClause + `
	ORDER BY t.id, p.time`, queryParams...)
	if err != nil {
		log.Errorf("Can not load track list: %v", err)
		return []Track{}
	}
	defer rows.Close()

	results := make([]Track, 0)
	var current Track
	var prevTrackId int64 = -1

	for rows.Next() {
		var id int64
		var title string
		var pathStr string
		var t_type string
		var p_id int64
		var p_type string
		var p_title string
		var p_content string
		var p_pointStr string
		var p_time time.Time

		err := rows.Scan(&id, &title, &pathStr, &t_type, &p_id, &p_type, &p_title, &p_content, &p_pointStr, &p_time)
		if err != nil {
			log.Fatal(err)
		}

		if prevTrackId != id {

			if (prevTrackId > 0) {
				results = append(results, current)
			}

			var path LineString
			err := json.Unmarshal([]byte(pathStr), &path)
			if err != nil {
				log.Errorf("Can not parse path for track %s: %v", path, err)
				continue
			}

			trackType, err := parseTrackType(t_type)
			if err != nil {
				log.Errorf("Can not parse track type for track %s: %v", id, err.Error())
				continue
			}

			current = Track{
				Id:id,
				Title:title,
				Points:make([]EventPoint, 0),
				Path: path.Coordinates,
				Type: trackType,
			}
			prevTrackId = id
		}

		if p_id < 0 {
			continue
		}

		var point PgPoint
		err = json.Unmarshal([]byte(p_pointStr), &point)
		if err != nil {
			log.Errorf("Can not parse point for track %s: %v", p_pointStr, err)
			continue
		}
		eventPointType, err := parseEventPointType(p_type)
		if err != nil {
			log.Errorf("Can not parse point type for track %s: %v", p_pointStr, err)
			continue
		}

		current.Points = append(current.Points, EventPoint{
			Id:p_id,
			Type:eventPointType,
			Title:p_title,
			Content:p_content,
			Time: JSONTime(p_time),
			Point: point.Coordinates,
		})
	}

	if (prevTrackId > 0) {
		results = append(results, current)
	}

	return results
}

func (s postgresStorage) AddTracks(tracks ...Track) error {
	log.Info("Inserting tracks")
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO track(title, path, type) VALUES ($1, ST_GeomFromGeoJSON($2), $3)")
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
		r, err := stmt.Exec(track.Title, string(pathBytes), string(track.Type))
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
		return fmt.Errorf("Updated %d rows for event point %d", rowsAffected, track.Id)
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

type PgPoint struct {
	Coordinates Point `json:"coordinates"`
}
