package dao

import (
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/lib/pq"
)

func NewSrtmPostgresDao(postgresStorage PostgresStorage) SrtmDao {
	return srtmPostgresStorage{
		PostgresStorage:   postgresStorage,
		selectRasterQuery: "SELECT array_to_json(ST_DumpValues(rast, 1, true))::text from srtm WHERE lat=$1 AND lon=$2",
	}
}

type srtmPostgresStorage struct {
	PostgresStorage
	selectRasterQuery string
}

func (this srtmPostgresStorage) GetRaster(lat int, lon int) (geo.Bytearea2D, bool, error) {
	raster, found, err := this.PostgresStorage.DoFindAndReturn(this.selectRasterQuery, rasterMapper, lat, lon)
	if err != nil || !found {
		return nil, found, err
	}
	return raster.(geo.Bytearea2D), found, err
}

func rasterMapper(rows *sql.Rows) (geo.Bytearea2D, error) {
	var jsonRaster string
	err := rows.Scan(&jsonRaster)
	if err != nil {
		return nil, err
	}

	data := make([][]int32, 3601)
	err = json.Unmarshal([]byte(jsonRaster), &data)
	if err != nil {
		return nil, err
	}

	return geo.InitBytearea2D(data)
}

func (this waterWayStorage) PathHeightPersister() (PathHeightPersister, error) {
	stmt, err := this.db.Prepare(this.updatePathHeightAndDistQuery)
	if err != nil {
		return PathHeightPersisterImpl{}, err
	}
	return PathHeightPersisterImpl{
		stmt: stmt,
	}, nil
}

type PathHeightPersisterImpl struct {
	stmt *sql.Stmt
}

func (this PathHeightPersisterImpl) Add(id int64, pathHeight []int32, dists []float64) error {
	_, err := this.stmt.Exec(id, pq.Array(pathHeight), pq.Array(dists))
	return err
}

func (this PathHeightPersisterImpl) Close() error {
	return this.stmt.Close()
}
