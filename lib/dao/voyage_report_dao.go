package dao

import (
	"database/sql"
	"time"
	"github.com/lib/pq"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/dao/queries"
)

func NewVoyageReportPostgresDao(postgresStorage PostgresStorage) VoyageReportDao {
	return voyageReportStorage{PostgresStorage:postgresStorage,
		upsertQuery: queries.SqlQuery("voyage-report", "upsert"),
		getLastIdQuery: queries.SqlQuery("voyage-report", "get-last-id"),
		listQuery: queries.SqlQuery("voyage-report", "list"),
		upsertRiverLinkQuery: queries.SqlQuery("voyage-report", "upsert-river-link"),
	}
}

type voyageReportStorage struct {
	PostgresStorage
	upsertQuery          string
	getLastIdQuery       string
	listQuery            string
	upsertRiverLinkQuery string
}

func (this voyageReportStorage) UpsertVoyageReports(reports ...VoyageReport) ([]VoyageReport, error) {
	reports_i := make([]interface{}, len(reports))
	for i := 0; i < len(reports); i++ {
		reports_i[i] = reports[i]
	}
	ids, err := this.updateReturningId(this.upsertQuery,
		func(entity interface{}) ([]interface{}, error) {
			_report := entity.(VoyageReport)
			tags, err := json.Marshal(_report.Tags)
			if err != nil {
				return []interface{}{}, err
			}
			return []interface{}{_report.Title, _report.RemoteId, _report.Source, _report.Url, _report.DatePublished, _report.DateModified, _report.DateOfTrip, tags}, nil
		}, reports_i...)

	if err != nil {
		return []VoyageReport{}, err
	}

	result := make([]VoyageReport, len(reports))
	copy(result, reports)
	for i := 0; i < len(reports); i++ {
		result[i].Id = ids[i]
	}
	return result, nil
}

func (this voyageReportStorage) GetLastId(source string) (interface{}, error) {
	lastDate := time.Unix(0, 0)
	_, err := this.doFind(this.getLastIdQuery, func(rows *sql.Rows) error {
		rows.Scan(&lastDate)
		return nil
	}, source)
	if err != nil {
		return nil, err
	}
	return lastDate, nil
}

func (this voyageReportStorage) AssociateWithRiver(voyageReportId, riverId int64) error {
	return this.performUpdates(this.upsertRiverLinkQuery,
		func(entity interface{}) ([]interface{}, error) {
			return entity.([]interface{}), nil
		}, []interface{}{voyageReportId, riverId})
}

func (this voyageReportStorage) List(riverId int64, limitByGroup int) ([]VoyageReport, error) {
	dateOfTrip := pq.NullTime{}
	result, err := this.doFindList(this.listQuery, func(rows *sql.Rows) (VoyageReport, error) {
		report := VoyageReport{}
		var rNum int
		var tags string
		err := rows.Scan(&rNum, &report.Id, &report.Title, &report.RemoteId, &report.Source, &report.Url, &report.DatePublished, &report.DateModified, &dateOfTrip, &tags)
		if dateOfTrip.Valid {
			report.DateOfTrip = dateOfTrip.Time
		}
		err = json.Unmarshal([]byte(tags), &report.Tags)
		if err != nil {
			return report, err
		}
		return report, err

	}, riverId, limitByGroup)

	if err != nil {
		return []VoyageReport{}, err
	}
	return result.([]VoyageReport), nil
}
