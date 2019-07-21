package dao

import (
	"encoding/json"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/geo"
)

func NewWaterWayOsmRefPostgresDao(postgresStorage PostgresStorage) WaterWayOsmRefDao {
	return waterWayOsmRefStorage{
		PostgresStorage: postgresStorage,
		insertQuery:     queries.SqlQuery("water-way-osm-ref", "insert"),
	}
}

type waterWayOsmRefStorage struct {
	PostgresStorage
	insertQuery string
}

func (this waterWayOsmRefStorage) Insert(refs ...WaterWayOsmRef) error {
	vars := make([]interface{}, len(refs))
	for i, r := range refs {
		vars[i] = r
	}

	return this.performUpdates(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		ref := entity.(WaterWayOsmRef)
		pointBytes, err := json.Marshal(geo.NewPgGeoPoint(ref.CrossPoint))
		if err != nil {
			return []interface{}{}, err
		}
		return []interface{}{ref.Id, ref.RefId, string(pointBytes)}, nil
	}, vars...)
}
