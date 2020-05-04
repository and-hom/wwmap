package dao

import (
	"database/sql"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/and-hom/wwmap/lib/util"
	"sort"
)

func NewTransferPostgresDao(postgresStorage PostgresStorage) TransferDao {
	return transferStorage{
		PostgresStorage:        postgresStorage,
		listQuery:              queries.SqlQuery("transfer", "list"),
		listByRiverQuery:       queries.SqlQuery("transfer", "list-by-river"),
		listFullQuery:          queries.SqlQuery("transfer", "list-full"),
		insertQuery:            queries.SqlQuery("transfer", "insert"),
		updateQuery:            queries.SqlQuery("transfer", "update"),
		deleteQuery:            queries.SqlQuery("transfer", "delete"),
		listRefsByRiverQuery:   queries.SqlQuery("transfer", "list-refs-by-river"),
		insertRefsQuery:        queries.SqlQuery("transfer", "insert-refs"),
		deleteRefsQuery:        queries.SqlQuery("transfer", "delete-refs"),
		deleteRefsByRiverQuery: queries.SqlQuery("transfer", "delete-refs-by-river"),
	}
}

type transferStorage struct {
	PostgresStorage
	listQuery              string
	listByRiverQuery       string
	listFullQuery          string
	insertQuery            string
	updateQuery            string
	deleteQuery            string
	insertRefsQuery        string
	deleteRefsQuery        string
	deleteRefsByRiverQuery string
	listRefsByRiverQuery   string
}

func (this transferStorage) List() ([]Transfer, error) {
	result, err := this.DoFindList(this.listQuery, transferMapper)
	if err != nil {
		return []Transfer{}, err
	}
	return result.([]Transfer), nil
}

func (this transferStorage) ByRiver(riverId int64) ([]Transfer, error) {
	result, err := this.DoFindList(this.listByRiverQuery, transferMapper, riverId)
	if err != nil {
		return []Transfer{}, err
	}
	return result.([]Transfer), nil
}

func (this transferStorage) ListFull() ([]TransferFull, error) {
	rows, err := this.db.Query(this.listFullQuery)
	if err != nil {
		return nil, err
	}
	defer util.DeferCloser(rows)

	byId := make(map[int64]TransferFull)

	for rows.Next() {
		t := TransferFull{}
		stationsString := ""
		var riverId sql.NullInt64
		var regionId sql.NullInt64
		var countryId sql.NullInt64
		var riverTitle sql.NullString
		var bounds sql.NullString

		if err := rows.Scan(&t.Id, &t.Title, &stationsString, &t.Description, &riverId, &regionId, &countryId, &riverTitle, &bounds); err != nil {
			return nil, err
		}

		var river *TransferToRiver = nil
		if riverId.Valid {
			river = &TransferToRiver{IdTitle{riverId.Int64, riverTitle.String}, regionId.Int64, countryId.Int64, nil}
			if bounds.Valid {
				b, err := ParseBounds(bounds.String)
				if err != nil {
					return nil, err
				}
				river.Bounds = &b
			}
		}

		existing, found := byId[t.Id]
		if found && river != nil {
			existing.Rivers = append(existing.Rivers, *river)
			byId[t.Id] = existing
		} else if !found {
			if err := json.Unmarshal([]byte(stationsString), &t.Stations); err != nil {
				return nil, err
			}
			if river != nil {
				t.Rivers = []TransferToRiver{*river}
			} else {
				t.Rivers = []TransferToRiver{}
			}
			byId[t.Id] = t
		}
	}

	result := make([]TransferFull, 0, len(byId))
	for _, v := range byId {
		result = append(result, v)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Id < result[j].Id
	})
	return result, nil
}

func (this transferStorage) Insert(transfer TransferFull) (int64, error) {
	id, err := this.UpdateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		_e := entity.(TransferFull)
		stations, err := json.Marshal(_e.Stations)
		if err != nil {
			return nil, err
		}
		return []interface{}{_e.Title, string(stations), _e.Description}, nil
	}, true, transfer)
	if err != nil {
		return 0, err
	}
	resultId := id[0]

	r := make([]interface{}, len(transfer.Rivers))
	for i := 0; i < len(transfer.Rivers); i++ {
		r[i] = transfer.Rivers[i]
	}
	if err := this.PerformUpdates(this.insertRefsQuery, func(entity interface{}) ([]interface{}, error) {
		r := entity.(TransferToRiver)
		return []interface{}{resultId, r.Id}, nil
	}, r...); err != nil {
		return 0, err
	}

	return resultId, err
}

func (this transferStorage) Update(transfer TransferFull) error {
	return this.PostgresStorage.WithinTx(func(tx interface{}) error {
		if err := this.PerformUpdatesWithinTxOptionally(tx, this.updateQuery, func(entity interface{}) ([]interface{}, error) {
			_e := entity.(TransferFull)
			stations, err := json.Marshal(_e.Stations)
			if err != nil {
				return nil, err
			}
			return []interface{}{_e.Id, _e.Title, string(stations), _e.Description}, nil
		}, transfer); err != nil {
			return err
		}

		if err := this.PerformUpdatesWithinTxOptionally(tx, this.deleteRefsQuery, IdMapper, transfer.Id); err != nil {
			return err
		}

		r := make([]interface{}, len(transfer.Rivers))
		for i := 0; i < len(transfer.Rivers); i++ {
			r[i] = transfer.Rivers[i]
		}
		if err := this.PerformUpdatesWithinTxOptionally(tx, this.insertRefsQuery, func(entity interface{}) ([]interface{}, error) {
			r := entity.(TransferToRiver)
			return []interface{}{transfer.Id, r.Id}, nil
		}, r...); err != nil {
			return err
		}
		return nil
	})
}

func (this transferStorage) Remove(id int64) error {
	return this.PostgresStorage.WithinTx(func(tx interface{}) error {
		if err := this.PerformUpdatesWithinTxOptionally(tx, this.deleteRefsQuery, IdMapper, id); err != nil {
			return err
		}
		if err := this.PerformUpdatesWithinTxOptionally(tx, this.deleteQuery, IdMapper, id); err != nil {
			return err
		}
		return nil
	})
}

func (this transferStorage) SetLinksForRiver(riverId int64, transfers []int64) error {
	return this.PostgresStorage.WithinTx(func(tx interface{}) error {
		if err := this.PerformUpdatesWithinTxOptionally(tx, this.deleteRefsByRiverQuery, IdMapper, riverId); err != nil {
			return err
		}

		r := make([]interface{}, len(transfers))
		for i := 0; i < len(transfers); i++ {
			r[i] = []interface{}{transfers[i], riverId}
		}
		if err := this.PerformUpdatesWithinTxOptionally(tx, this.insertRefsQuery, ArrayMapper, r...); err != nil {
			return err
		}

		return nil
	})
}

func (this transferStorage) GetIdsForRiver(riverId int64) ([]int64, error) {
	result, err := this.DoFindList(this.listRefsByRiverQuery, Int64ColumnMapper, riverId)
	if err != nil {
		return []int64{}, err
	}
	return result.([]int64), nil
}

func transferMapper(rows *sql.Rows) (Transfer, error) {
	transfer := Transfer{}
	stationsString := ""

	if err := rows.Scan(&transfer.Id, &transfer.Title, &stationsString, &transfer.Description); err != nil {
		return transfer, err
	}
	err := json.Unmarshal([]byte(stationsString), &transfer.Stations)
	return transfer, err
}
