package dao

import (
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewWaterWayRefPostgresDao(postgresStorage PostgresStorage) WaterWayRefDao {
	return waterWayRefStorage{
		PostgresStorage: postgresStorage,
		selectAllQuery:  queries.SqlQuery("water-way-ref", "all"),
	}
}

type waterWayRefStorage struct {
	PostgresStorage
	selectAllQuery string
}

func (this waterWayRefStorage) RefsById() (map[int64][]int64, error) {
	result := make(map[int64][]int64)
	rows, err := this.db.Query(this.selectAllQuery)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	prevId := int64(-1)
	refIds := make([]int64, 0, 10)
	for rows.Next() {
		var id int64
		var refId int64
		err := rows.Scan(&id, &refId)
		if err != nil {
			return result, err
		}
		if prevId > 0 && id != prevId {
			result[prevId] = refIds
			refIds = make([]int64, 0, 10)
			prevId = id
		}
		refIds = append(refIds, refId)
	}
	if len(refIds) > 0 {
		result[prevId] = refIds
	}
	return result, nil
}
