package main

import (
	"time"
	log "github.com/Sirupsen/logrus"
	"database/sql"
	_ "github.com/lib/pq"
	"encoding/json"
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
	getTracks(bbox Bbox) TrackList
	insert(track ...Track) error
}

type DummyStorage struct {

}

func (s DummyStorage) getTracks(bbox Bbox) TrackList {
	return []Track{
		Track{
			Id: int64(1),
			Title:"Track1",
			Path:[]Point{
				Point{x:56.2877096985583, y:37.5007003462651, },
				Point{x:56.2877096985583, y:37.5002282774785, },
				Point{x:56.2881384381127, y:37.4998420393804, },
				Point{x:56.2891388117074, y:37.4973529494145, },
				Point{x:56.2919968777318, y:37.4951642668584, },
				Point{x:56.2927351767222, y:37.4951642668584, },
				Point{x:56.2936163536233, y:37.4946063673833, },
				Point{x:56.2953310180125, y:37.4946063673833, },
				Point{x:56.2933781997074, y:37.4879544890264, },
				Point{x:56.2940450269324, y:37.4876111662724, },
				Point{x:56.2984267449926, y:37.4852937376835, },
				Point{x:56.3024746249771, y:37.481045118604, },
				Point{x:56.3051174210983, y:37.4796289122441, },
				Point{x:56.3029746274536, y:37.468084684644, },
				Point{x:56.3015222212059, y:37.4659818327763, },
				Point{x:56.3020222361459, y:37.4649518645146, },
				Point{x:56.3011650636713, y:37.4585574782231, },
				Point{x:56.3006888484319, y:37.457999578748, },
				Point{x:56.2953548322541, y:37.4449533140996, },
				Point{x:56.2928780716538, y:37.4410909331182, },
				Point{x:56.2923541209595, y:37.4321645415166, },
				Point{x:56.2900200715184, y:37.4254697478154, },
				Point{x:56.2890197209855, y:37.4243539488652, },
				Point{x:56.2865425498548, y:37.4147409117558, },
				Point{x:56.2832076429274, y:37.408646932874, },
				Point{x:56.2801583306034, y:37.4006646788457, },
				Point{x:56.2767752152185, y:37.394055715833, },
				Point{x:56.2739160111405, y:37.3896783507207, },
				Point{x:56.2730105519558, y:37.3871892607549, },
				Point{x:56.2716521951454, y:37.3839276945929, },
				Point{x:56.269025467908, y:37.3744434035162, },
				Point{x:56.2663985595555, y:37.379807821546, },
				Point{x:56.2638908871682, y:37.379979482923, },
				Point{x:56.2619323996772, y:37.3777908003668, },
				Point{x:56.2608814624543, y:37.3758596098761, },
				Point{x:56.26006935475, y:37.374400488172, },
				Point{x:56.2588511607457, y:37.3730701125006, },
				Point{x:56.2574179414771, y:37.3709672606329, },
				Point{x:56.2562713272581, y:37.3680490172247, },
				Point{x:56.2560802215348, y:37.3662465727667, },
				Point{x:56.2551246785458, y:37.3645299589972, },
				Point{x:56.2540257744953, y:37.3661607420782, },

			},
			Points:[]EventPoint{
				EventPoint{Point{x:56.2877096985583, y:37.5007003462651, },
					2, "Старт", "Начало нашего маршрута", JSONTime(time.Now())},
				EventPoint{Point{x:56.26006935475, y:37.374400488172, },
					3, "Фигня какая-то", "Из леса вышел лосось", JSONTime(time.Now())},
			},
		},
	}
}

func (s DummyStorage) insert(track ...Track) error {
	return nil
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

func (s postgresStorage) getTracks(bbox Bbox) TrackList {
	rows, err := s.db.Query(`SELECT t.id, t.title, ST_AsGeoJSON(t.path) as path,
	COALESCE(p.id, -1), COALESCE(p.title,''), COALESCE(p.text,''),
	COALESCE(ST_AsGeoJSON(p.point),'') as point, COALESCE(p.time, now())
	FROM track t LEFT OUTER JOIN point p ON t.id=p.track_id
	WHERE path && ST_MakeEnvelope($1,$2,$3,$4)
	ORDER BY t.id, p.time`, bbox.X1, bbox.Y1, bbox.X2, bbox.Y2)
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
		var p_id int64
		var p_title string
		var p_text string
		var p_pointStr string
		var p_time time.Time

		err := rows.Scan(&id, &title, &pathStr, &p_id, &p_title, &p_text, &p_pointStr, &p_time)
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

			current = Track{
				Id:id,
				Title:title,
				Points:make([]EventPoint, 0),
				Path: path.Coordinates,
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
		current.Points = append(current.Points, EventPoint{
			Id:p_id,
			Title:p_title,
			Text:p_text,
			Time: JSONTime(p_time),
			Point: point.Coordinates,
		})
	}

	if (prevTrackId > 0) {
		results = append(results, current)
	}

	return results
}

func (s postgresStorage) insert(tracks ...Track) error {
	log.Info("Inserting tracks")
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO track(title, path) VALUES ($1, ST_GeomFromGeoJSON($2))")
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
		r, err := stmt.Exec(track.Title, string(pathBytes))
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

type PgPoint struct {
	Coordinates Point `json:"coordinates"`
}
