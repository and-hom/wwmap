package dao

import (
	"database/sql"
	"github.com/lib/pq"
	"fmt"
)

type ReportStorage struct {
	PostgresStorage
}

func (this ReportStorage) AddReport(report Report) error {
	_, err := this.insertReturningId("INSERT INTO report(object_id,comment) VALUES($1,$2) RETURNING id", report.ObjectId, report.Comment)
	return err;
}

func (this ReportStorage) ListUnread(limit int) ([]Report, error) {
	reports, err := this.doFindList(
		"SELECT id, object_id, comment, created_at FROM report WHERE NOT read ORDER BY created_at ASC LIMIT $1",
		func(rows *sql.Rows) (Report, error) {
			report := Report{}
			objectId := sql.NullInt64{}
			err := rows.Scan(&report.Id, &objectId, &report.Comment, &report.CreatedAt)
			if err != nil {
				return Report{}, err
			}
			if objectId.Valid {
				report.ObjectId = objectId.Int64
			}
			return report, nil
		}, limit)
	return reports.([]Report), err
}

func (this ReportStorage) MarkRead(reports []int64) error {
	fmt.Println(reports)
	return this.performUpdates(
		"UPDATE report SET read=TRUE WHERE id = ANY($1)",
		func(ids interface{}) ([]interface{}, error) {
			return []interface{}{ids}, nil;
		}, pq.Array(reports))
}