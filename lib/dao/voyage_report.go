package dao

import (
	"database/sql"
	"time"
	"github.com/lib/pq"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/dao/queries"
	log "github.com/Sirupsen/logrus"
)

func NewVoyageReportPostgresDao(postgresStorage PostgresStorage) VoyageReportDao {
	return voyageReportStorage{PostgresStorage:postgresStorage,
		upsertQuery: queries.SqlQuery("voyage-report", "upsert"),
		getLastIdQuery: queries.SqlQuery("voyage-report", "get-last-id"),
		listQuery: queries.SqlQuery("voyage-report", "list"),
		upsertRiverLinkQuery: queries.SqlQuery("voyage-report", "upsert-river-link"),
		listAllQuery: queries.SqlQuery("voyage-report", "list-all"),
		deleteRiverLinkQuery: queries.SqlQuery("voyage-report", "delete-river-link"),
	}
}

type voyageReportStorage struct {
	PostgresStorage
	upsertQuery          string
	getLastIdQuery       string
	listQuery            string
	listAllQuery         string
	upsertRiverLinkQuery string
	deleteRiverLinkQuery string
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
			return []interface{}{_report.Title, _report.RemoteId, _report.Source, _report.Url, _report.DatePublished, _report.DateModified, _report.DateOfTrip, tags, _report.Author}, nil
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
	lastDate, found, err := this.doFindAndReturn(this.getLastIdQuery, func(rows *sql.Rows) (time.Time, error) {
		lastDate := time.Unix(0, 0)
		err := rows.Scan(&lastDate)
		return lastDate, err
	}, source)
	if err != nil {
		return nil, err
	}
	if !found {
		return time.Unix(0, 0), nil
	}
	return lastDate, nil
}

func (this voyageReportStorage) ForEach(source string, callback func(report *VoyageReport) error) error {
	return this.forEach(this.listAllQuery,
		func(rows *sql.Rows) error {
			report, err := readReportFromRows(rows)
			if err != nil {
				return err
			}
			return callback(&report)
		}, source)
}

func (this voyageReportStorage) AssociateWithRiver(voyageReportId, riverId int64) error {
	return this.performUpdates(this.upsertRiverLinkQuery, ArrayMapper, []interface{}{voyageReportId, riverId})
}

func (this voyageReportStorage) List(riverId int64, limitByGroup int) ([]VoyageReport, error) {
	result, err := this.doFindList(this.listQuery, readReportFromRows, riverId, limitByGroup)

	if err != nil {
		return []VoyageReport{}, err
	}
	return result.([]VoyageReport), nil
}

func (this voyageReportStorage) RemoveRiverLink(id int64, tx interface{}) error {
	log.Infof("Remove spot %d", id)
	return this.performUpdatesWithinTxOptionally(tx, this.deleteRiverLinkQuery, IdMapper, id)
}

func readReportFromRows(rows *sql.Rows) (VoyageReport, error) {
	report := VoyageReport{}
	dateOfTrip := pq.NullTime{}
	var tags string
	err := rows.Scan(&report.Id, &report.Title, &report.RemoteId, &report.Source, &report.Url, &report.DatePublished, &report.DateModified, &dateOfTrip, &tags, &report.Author)
	if dateOfTrip.Valid {
		report.DateOfTrip = dateOfTrip.Time
	}
	err = json.Unmarshal([]byte(tags), &report.Tags)
	if err != nil {
		return report, err
	}
	return report, err
}
