package dao

import (
	"database/sql"
	"fmt"
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewLevelSensorPostgresDao(postgresStorage PostgresStorage) LevelSensorDao {
	return &levelSensorStorage{
		PostgresStorage:         postgresStorage,
		listQuery:               queries.SqlQuery("level-sensor", "list"),
		byIdQuery:               queries.SqlQuery("level-sensor", "find"),
		setGraduationQuery:      queries.SqlQuery("level-sensor", "set-graduation"),
		checkAndCreateIfMissing: queries.SqlQuery("level-sensor", "check-and-create-if-missing"),
	}
}

type levelSensorStorage struct {
	PostgresStorage
	listQuery               string
	byIdQuery               string
	setGraduationQuery      string
	checkAndCreateIfMissing string
}

func (this levelSensorStorage) List() ([]LevelSensor, error) {
	lst, err := this.doFindList(this.listQuery, scanLevelSensor)
	if err != nil {
		return nil, err
	}
	return lst.([]LevelSensor), nil
}

func (this levelSensorStorage) Find(id string) (LevelSensor, error) {
	p, found, err := this.doFindAndReturn(this.byIdQuery, scanLevelSensor, id)
	if err != nil {
		return LevelSensor{}, err
	}
	if !found {
		return LevelSensor{}, fmt.Errorf("LevelSensor with id=%d not found", id)
	}
	return p.(LevelSensor), nil
}

func (this levelSensorStorage) SetGraduation(id string, graduation [LEVEL_GRADUATION]int) error {
	params := []interface{}{id}
	for i := 0; i < LEVEL_GRADUATION; i++ {
		params = append(params, graduation[i])
	}
	return this.performUpdates(this.setGraduationQuery, ArrayMapper, params)
}

func (this levelSensorStorage) CreateIfMissing(id string) error {
	return this.performUpdates(this.checkAndCreateIfMissing, IdMapper, id)
}

func scanLevelSensor(rows *sql.Rows) (LevelSensor, error) {
	result := LevelSensor{}
	err := rows.Scan(&result.Id, &result.L[0], &result.L[1], &result.L[2], &result.L[3])
	return result, err
}
