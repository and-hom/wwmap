package dao

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/chai2010/tiff"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func NewSrtmPostgresDao(postgresStorage PostgresStorage) SrtmDao {
	return srtmPostgresStorage{
		PostgresStorage:   postgresStorage,
		selectRasterQuery: "SELECT ST_AsTIFF(rast) from srtm WHERE lat=$1 AND lon=$2",
		listPointsQuery:   "SELECT lat, lon FROM srtm WHERE srtm.lon >= $1 AND srtm.lat >= $2 AND srtm.lon <= $3 AND srtm.lat <= $4",
	}
}

type srtmPostgresStorage struct {
	PostgresStorage
	selectRasterQuery string
	listPointsQuery   string
}

func (this srtmPostgresStorage) GetRaster(lat int, lon int) (geo.Bytearea2D, bool, error) {
	raster, found, err := this.PostgresStorage.DoFindAndReturn(this.selectRasterQuery, rasterMapper, lat, lon)
	if err != nil || !found {
		return nil, found, err
	}
	return raster.(geo.Bytearea2D), found, err
}

func (this srtmPostgresStorage) GetRasterCoords(bbox geo.Bbox) ([]geo.PointInt, error) {
	lst, err := this.PostgresStorage.DoFindList(this.listPointsQuery, scanPoint, int(bbox.Y1), int(bbox.X1), int(bbox.Y2), int(bbox.X2))
	if err != nil {
		return []geo.PointInt{}, err
	}
	return lst.([]geo.PointInt), nil
}

func scanPoint(rows *sql.Rows) (geo.PointInt, error) {
	result := geo.PointInt{}
	err := rows.Scan(&result.Y, &result.X)
	return result, err
}

func rasterMapper(rows *sql.Rows) (geo.Bytearea2D, error) {
	var rasterData []byte

	err := rows.Scan(&rasterData)
	if err != nil {
		return nil, err
	}

	m, errs, err := tiff.DecodeAll(bytes.NewReader(rasterData))
	if err != nil {
		log.Println(err)
	}

	if len(errs) > 0 && len(errs[0]) > 0 && errs[0][0] != nil {
		return nil, errs[0][0]
	}

	if len(m) < 1 || len(m[0]) < 1 {
		return nil, fmt.Errorf("No raster loaded!")
	}

	return geo.InitImageBasedBytearea2D(m[0][0])
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

func (this PathHeightPersisterImpl) Add(id int64, pathHeight []int, dists []float64) error {
	_, err := this.stmt.Exec(id, pq.Array(pathHeight), pq.Array(dists))
	return err
}

func (this PathHeightPersisterImpl) Close() error {
	return this.stmt.Close()
}
