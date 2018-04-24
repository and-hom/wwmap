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

func (this ReportStorage) ListUnread(limit int) ([]ReportWithName, error) {
	reports, err := this.doFindList(
		"SELECT report.id, COALESCE(white_water_rapid.id, -1) as title, COALESCE(white_water_rapid.title, '') as title, COALESCE(river.title, '') as river_title, report.comment, report.created_at " +
			"FROM report LEFT OUTER JOIN white_water_rapid ON report.object_id=white_water_rapid.id " +
			"LEFT OUTER JOIN river ON white_water_rapid.river_id=river.id " +
			"WHERE NOT report.read " +
			"ORDER BY created_at ASC LIMIT $1",
		func(rows *sql.Rows) (ReportWithName, error) {
			report := ReportWithName{}
			err := rows.Scan(&report.Id, &report.ObjectId, &report.Title, &report.RiverTitle, &report.Comment, &report.CreatedAt)
			if err != nil {
				return ReportWithName{}, err
			}
			return report, nil
		}, limit)
	return reports.([]ReportWithName), err
}

func (this ReportStorage) MarkRead(reports []int64) error {
	fmt.Println(reports)
	return this.performUpdates(
		"UPDATE report SET read=TRUE WHERE id = ANY($1)",
		func(ids interface{}) ([]interface{}, error) {
			return []interface{}{ids}, nil;
		}, pq.Array(reports))
}