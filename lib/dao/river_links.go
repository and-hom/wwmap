package dao

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

type riverLinksStorage struct {
	PostgresStorage
	insertRefsQuery        string
	deleteRefsByRiverQuery string
	listRefsByRiverQuery   string
	deleteRefsQuery        string
	listRivers             string
}

func (this riverLinksStorage) GetIdsForRiver(riverId int64) ([]int64, error) {
	result, err := this.DoFindList(this.listRefsByRiverQuery, Int64ColumnMapper, riverId)
	if err != nil {
		return []int64{}, err
	}
	return result.([]int64), nil
}

func (this riverLinksStorage) SetLinksForRiver(riverId int64, refIds []int64) error {
	return this.PostgresStorage.WithinTx(func(tx interface{}) error {
		if err := this.PerformUpdatesWithinTxOptionally(tx, this.deleteRefsByRiverQuery, IdMapper, riverId); err != nil {
			return err
		}

		r := make([]interface{}, len(refIds))
		for i := 0; i < len(refIds); i++ {
			r[i] = []interface{}{refIds[i], riverId}
		}
		if err := this.PerformUpdatesWithinTxOptionally(tx, this.insertRefsQuery, ArrayMapper, r...); err != nil {
			return err
		}

		return nil
	})
}

func (this riverLinksStorage) loadRivers(ids []int64) (map[int64]LinkedEntityRiver, error) {
	riversById := make(map[int64]LinkedEntityRiver)
	if len(ids) == 0 {
		return riversById, nil
	}
	result, err := this.DoFindList(this.listRivers, riversMapper, pq.Int64Array(ids))
	if err != nil {
		return nil, err
	}
	rivers := result.([]LinkedEntityRiver)
	for i := 0; i < len(rivers); i++ {
		riversById[rivers[i].Id] = rivers[i]
	}
	return riversById, nil
}

func riversMapper(rows *sql.Rows) (LinkedEntityRiver, error) {
	river := LinkedEntityRiver{}
	var bounds sql.NullString

	if err := rows.Scan(&river.Id, &river.RegionId, &river.CountryId, &river.Title, &bounds); err != nil {
		return river, err
	}

	if bounds.Valid {
		b, err := ParseBounds(bounds.String)
		if err != nil {
			return river, err
		}
		river.Bounds = &b
	}

	return river, nil
}

func (this riverLinksStorage) enrichWithRiverData(entities []ILinkedEntity) error {
	riverIdsMap := make(map[int64]bool)

	for _, entity := range entities {
		for _, id := range entity.GetRivers() {
			riverIdsMap[id] = true
		}
	}
	riverIds := make([]int64, 0, len(riverIdsMap))
	for id, _ := range riverIdsMap {
		riverIds = append(riverIds, id)
	}

	rivers, err := this.loadRivers(riverIds)
	if err != nil {
		return err
	}

	for _, t := range entities {
		riverIds := t.GetRivers()
		if len(riverIds) == 0 {
			continue
		}

		riverData := make([]LinkedEntityRiver, 0, len(riverIds))
		for i := 0; i < len(riverIds); i++ {
			riverId := riverIds[i]
			r, f := rivers[riverId]
			if f {
				riverData = append(riverData, r)
			}
		}

		fmt.Println(riverData)
		t.SetRiversData(&riverData)
	}
	return nil
}

func (this riverLinksStorage) update(updateQuery string, entityId int64, rivers []int64, fields []interface{}) error {
	return this.PostgresStorage.WithinTx(func(tx interface{}) error {
		if err := this.PerformUpdatesWithinTxOptionally(tx, updateQuery, ArrayMapper, fields); err != nil {
			return err
		}

		if err := this.PerformUpdatesWithinTxOptionally(tx, this.deleteRefsQuery, IdMapper, entityId); err != nil {
			return err
		}

		r := make([]interface{}, len(rivers))
		for i := 0; i < len(rivers); i++ {
			r[i] = rivers[i]
		}
		if err := this.PerformUpdatesWithinTxOptionally(tx, this.insertRefsQuery, func(entity interface{}) ([]interface{}, error) {
			riverId := entity.(int64)
			return []interface{}{entityId, riverId}, nil
		}, r...); err != nil {
			return err
		}
		return nil
	})
}
