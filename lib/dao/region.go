package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func NewRegionPostgresDao(postgresStorage PostgresStorage) RegionDao {
	return regionStorage{
		PostgresStorage:         postgresStorage,
		PropsManager:            PropertyManagerImpl{table: queries.SqlQuery("region", "table"), dao: &postgresStorage},
		listQuery:               queries.SqlQuery("region", "list-real"),
		listAllWithCountryQuery: queries.SqlQuery("region", "list-all-with-country"),
		getByIdQuery:            queries.SqlQuery("region", "get-by-id"),
		getFakeQuery:            queries.SqlQuery("region", "get-fake"),
		createFakeQuery:         queries.SqlQuery("region", "create-fake"),
		insertQuery:             queries.SqlQuery("region", "insert"),
		updateQuery:             queries.SqlQuery("region", "update"),
		deleteQuery:             queries.SqlQuery("region", "delete"),
		deleteInCountryQuery:    queries.SqlQuery("region", "delete-in-country"),
		parentIds:               queries.SqlQuery("region", "parent-ids"),
	}
}

type regionStorage struct {
	PostgresStorage
	PropsManager            PropertyManager
	getByIdQuery            string
	listQuery               string
	listAllWithCountryQuery string
	getFakeQuery            string
	createFakeQuery         string
	insertQuery             string
	updateQuery             string
	deleteQuery             string
	deleteInCountryQuery    string
	parentIds               string
}

func (this regionStorage) List(countryId int64) ([]Region, error) {
	lst, err := this.DoFindList(this.listQuery, scanFunc, countryId)
	if err != nil {
		return []Region{}, err
	}
	return lst.([]Region), nil
}

func (this regionStorage) Get(id int64) (Region, bool, error) {
	result, found, err := this.DoFindAndReturn(this.getByIdQuery, scanFuncI, id)
	if err != nil {
		return Region{}, found, err
	}
	if !found {
		return Region{}, found, nil
	}
	return result.(Region), found, nil
}

func (this regionStorage) ListAllWithCountry() ([]RegionWithCountry, error) {
	lst, err := this.DoFindList(this.listAllWithCountryQuery, func(rows *sql.Rows) (RegionWithCountry, error) {
		result := RegionWithCountry{
			Country: Country{},
		}
		err := rows.Scan(
			&result.Id,
			&result.Country.Id,
			&result.Country.Title,
			&result.Country.Code,
			&result.Title,
			&result.Fake,
			)
		return result, err
	})
	if err != nil {
		return []RegionWithCountry{}, err
	}
	return lst.([]RegionWithCountry), nil
}

func (this regionStorage) Props() PropertyManager {
	return this.PropsManager
}

func (this regionStorage) GetFake(countryId int64) (Region, bool, error) {
	result, found, err := this.DoFindAndReturn(this.getFakeQuery, scanFuncI, countryId)
	if err != nil {
		return Region{}, false, err
	}
	if !found {
		return Region{}, false, nil
	}
	return result.(Region), true, nil
}

func (this regionStorage) CreateFake(countryId int64) (int64, error) {
	ids, err := this.UpdateReturningId(this.createFakeQuery, IdMapper, true, countryId)
	if err != nil {
		return 0, err
	}
	return ids[0], err
}

func (this regionStorage) Save(regions ...Region) error {
	vars := make([]interface{}, len(regions))
	for i, p := range regions {
		vars[i] = p
	}
	return this.PerformUpdates(this.updateQuery, func(entity interface{}) ([]interface{}, error) {
		_region := entity.(Region)
		return []interface{}{_region.CountryId, _region.Title, _region.Fake, _region.Id}, nil
	}, vars...)
}

func (this regionStorage) Insert(region Region) (int64, error) {
	id, err := this.UpdateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(Region)
		return []interface{}{_e.CountryId, _e.Title, _e.Fake}, nil
	}, true, region)
	if err != nil {
		return 0, err
	}
	return id[0], err
}

func (this regionStorage) Remove(id int64) error {
	log.Infof("Remove region %d", id)
	return this.PerformUpdatesWithinTxOptionally(nil, this.deleteQuery, IdMapper, id)
}

func (this regionStorage) RemoveAllByCountry(countryId int64) error {
	log.Infof("Remove regions by country %d", countryId)
	return this.PerformUpdatesWithinTxOptionally(nil, this.deleteInCountryQuery, IdMapper, countryId)
}

func (this regionStorage) GetParentIds(regionIds []int64) (map[int64]RegionParentIds, error) {
	result := make(map[int64]RegionParentIds)

	_, err := this.DoFindList(this.parentIds, func(rows *sql.Rows) (int, error) {
		regionId := int64(0)
		parentIds := RegionParentIds{}
		err := rows.Scan(&regionId, &parentIds.CountryId, &parentIds.RegionTitle)

		if err == nil {
			result[regionId] = parentIds
		}
		return 0, err
	}, pq.Array(regionIds))

	if err != nil {
		return result, err
	}
	return result, nil
}

func scanFuncI(rows *sql.Rows) (interface{}, error) {
	return scanFunc(rows)
}

func scanFunc(rows *sql.Rows) (Region, error) {
	result := Region{}
	err := rows.Scan(&result.Id, &result.CountryId, &result.Title, &result.Fake)
	return result, err
}
