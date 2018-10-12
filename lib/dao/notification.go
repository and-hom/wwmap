package dao

import (
	"database/sql"
	"github.com/lib/pq"
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewNotificationPostgresDao(postgresStorage PostgresStorage) NotificationDao {
	return reportStorage{
		PostgresStorage : postgresStorage,
		insertQuery : queries.SqlQuery("notification", "insert"),
		listUnreadQuery : queries.SqlQuery("notification", "list-unread"),
		markReadQuery : queries.SqlQuery("notification", "mark-read"),
	}
}

type reportStorage struct {
	PostgresStorage
	insertQuery     string
	listUnreadQuery string
	markReadQuery   string
}

func (this reportStorage) Add(report Notification) error {
	_, err := this.updateReturningId(this.insertQuery, arrayMapper, []interface{}{report.ObjectId, report.Comment})
	return err;
}

func (this reportStorage) ListUnread(limit int) ([]ReportWithName, error) {
	reports, err := this.doFindList(this.listUnreadQuery,
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

func (this reportStorage) MarkRead(reports []int64) error {
	return this.performUpdates(this.markReadQuery,
		func(ids interface{}) ([]interface{}, error) {
			return []interface{}{ids}, nil;
		}, pq.Array(reports))
}