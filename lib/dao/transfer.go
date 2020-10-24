package dao

import (
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/lib/pq"
)

func NewTransferPostgresDao(postgresStorage PostgresStorage) TransferDao {
	return transferStorage{
		riverLinksStorage: riverLinksStorage{
			PostgresStorage:        postgresStorage,
			listRefsByRiverQuery:   queries.SqlQuery("transfer", "list-refs-by-river"),
			insertRefsQuery:        queries.SqlQuery("transfer", "insert-refs"),
			deleteRefsQuery:        queries.SqlQuery("transfer", "delete-refs"),
			deleteRefsByRiverQuery: queries.SqlQuery("transfer", "delete-refs-by-river"),
			listRivers:             queries.SqlQuery("linked-entity", "list-rivers"),
		},
		listQuery:        queries.SqlQuery("transfer", "list"),
		listByRiverQuery: queries.SqlQuery("transfer", "list-by-river"),
		insertQuery:      queries.SqlQuery("transfer", "insert"),
		updateQuery:      queries.SqlQuery("transfer", "update"),
		deleteQuery:      queries.SqlQuery("transfer", "delete"),
	}
}

type transferStorage struct {
	riverLinksStorage
	listQuery        string
	listByRiverQuery string
	listFullQuery    string
	insertQuery      string
	updateQuery      string
	deleteQuery      string
}

func (this transferStorage) List(withRivers bool) ([]Transfer, error) {
	result, err := this.DoFindList(this.listQuery, transferMapper)
	if err != nil {
		return []Transfer{}, err
	}

	transfers := result.([]Transfer)

	if withRivers {
		if err := this.enrichWithRiverData(convertTransfers(&transfers)); err != nil {
			return nil, err
		}
	}

	return transfers, nil
}

func convertTransfers(transfers *[]Transfer) []ILinkedEntity {
	result := make([]ILinkedEntity, len(*transfers))
	for i := 0; i < len(*transfers); i++ {
		result[i] = &(*transfers)[i]
	}
	return result
}

func (this transferStorage) ByRiver(riverId int64) ([]Transfer, error) {
	result, err := this.DoFindList(this.listByRiverQuery, transferMapper, riverId)
	if err != nil {
		return []Transfer{}, err
	}
	return result.([]Transfer), nil
}

func (this transferStorage) Insert(transfer Transfer) (int64, error) {
	stations, err := json.Marshal(transfer.Stations)
	if err != nil {
		return 0, err
	}
	fields := []interface{}{transfer.Title, string(stations), transfer.Description}
	return this.insert(this.insertQuery, transfer.Rivers, fields)
}

func (this transferStorage) Update(transfer Transfer) error {
	stations, err := json.Marshal(transfer.Stations)
	if err != nil {
		return err
	}
	fields := []interface{}{transfer.Id, transfer.Title, string(stations), transfer.Description}
	return this.update(this.updateQuery, transfer.Id, transfer.Rivers, fields)
}

func (this transferStorage) Remove(id int64) error {
	return this.PerformUpdates(this.deleteQuery, IdMapper, id)
}

func transferMapper(rows *sql.Rows) (Transfer, error) {
	transfer := Transfer{}
	stationsString := ""
	rivers := pq.Int64Array{}

	if err := rows.Scan(&transfer.Id, &transfer.Title, &stationsString, &transfer.Description, &rivers); err != nil {
		return transfer, err
	}
	transfer.Rivers = []int64(rivers)

	err := json.Unmarshal([]byte(stationsString), &transfer.Stations)
	return transfer, err
}
