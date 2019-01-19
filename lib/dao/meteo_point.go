package dao

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/geo"
	log "github.com/Sirupsen/logrus"
)

func NewMeteoPointPostgresDao(postgresStorage PostgresStorage) MeteoPointDao {
	return &meteoPointStorage{
		PostgresStorage: postgresStorage,
		insertQuery:     queries.SqlQuery("meteo-point", "insert"),
		listQuery:       queries.SqlQuery("meteo-point", "list"),
	}
}

type meteoPointStorage struct {
	PostgresStorage
	insertQuery string
	listQuery   string
}

func (this meteoPointStorage) Insert(entry MeteoPoint) error {
	_, err := this.updateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(MeteoPoint)
		pointBytes, err := json.Marshal(geo.NewPgGeoPoint(_e.Point))
		if err != nil {
			return nil, err
		}
		fmt.Println(string(pointBytes))
		return []interface{}{_e.Title, string(pointBytes)}, nil
	}, entry)
	return err
}

func (this meteoPointStorage) List() ([]MeteoPoint, error) {
	lst, err := this.doFindList(this.listQuery, func(rows *sql.Rows) (MeteoPoint, error) {
		result := MeteoPoint{}
		pointStr := ""

		err := rows.Scan(&result.Id, &result.Title, &pointStr)

		var pgPoint PgPoint
		err = json.Unmarshal([]byte(pointStr), &pgPoint)
		if err != nil {
			log.Errorf("Can not parse centroid point %s for meteo point %d: %v", pointStr, result.Id, err)
			return MeteoPoint{}, err
		}
		result.Point = pgPoint.GetPoint()

		return result, err
	})
	if err != nil {
		return []MeteoPoint{}, err
	}
	return lst.([]MeteoPoint), nil
}
