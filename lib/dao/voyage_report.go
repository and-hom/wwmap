package dao

import (
	"database/sql"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/lib/pq"
	"time"
)

func NewVoyageReportPostgresDao(postgresStorage PostgresStorage) VoyageReportDao {
	return voyageReportStorage{
		riverLinksStorage: riverLinksStorage{
			PostgresStorage:        postgresStorage,
			listRefsByRiverQuery:   queries.SqlQuery("voyage-report", "list-refs-by-river"),
			countRefsByRiverQuery:  queries.SqlQuery("voyage-report", "count-refs-by-river"),
			insertRefsQuery:        queries.SqlQuery("voyage-report", "insert-refs"),
			deleteRefsQuery:        queries.SqlQuery("voyage-report", "delete-refs"),
			deleteRefsByRiverQuery: queries.SqlQuery("voyage-report", "delete-refs-by-river"),
			listRivers:             queries.SqlQuery("linked-entity", "list-rivers"),
		},
		insertQuery:          queries.SqlQuery("voyage-report", "insert"),
		upsertQuery:          queries.SqlQuery("voyage-report", "upsert"),
		updateQuery:          queries.SqlQuery("voyage-report", "update"),
		getLastIdQuery:       queries.SqlQuery("voyage-report", "get-last-id"),
		findQuery:            queries.SqlQuery("voyage-report", "find"),
		listQuery:            queries.SqlQuery("voyage-report", "list"),
		listByRiverQuery:     queries.SqlQuery("voyage-report", "list-by-river"),
		listAllQuery:         queries.SqlQuery("voyage-report", "list-all"),
		removeQuery:          queries.SqlQuery("voyage-report", "remove"),
		upsertRiverLinkQuery: queries.SqlQuery("voyage-report", "upsert-river-link"),
		deleteRiverLinkQuery: queries.SqlQuery("voyage-report", "delete-river-link"),
	}
}

type voyageReportStorage struct {
	riverLinksStorage
	insertQuery          string
	upsertQuery          string
	updateQuery          string
	getLastIdQuery       string
	findQuery            string
	listQuery            string
	listByRiverQuery     string
	listAllQuery         string
	removeQuery          string
	upsertRiverLinkQuery string
	deleteRiverLinkQuery string
}

func (this voyageReportStorage) List(withRivers bool) ([]VoyageReport, error) {
	found, err := this.DoFindList(this.listQuery, readReportFromRows)
	if err != nil {
		return []VoyageReport{}, err
	}

	reports := found.([]VoyageReport)

	if withRivers {
		if err := this.enrichWithRiverData(convertVoyageReports(&reports)); err != nil {
			return nil, err
		}
	}

	return reports, nil
}

func (this voyageReportStorage) Insert(report VoyageReport) (int64, error) {
	fields, err := this.reportToArgs(false)(report)
	if err != nil {
		return 0, err
	}
	return this.insert(this.insertQuery, report.Rivers, fields)
}

func (this voyageReportStorage) Update(report VoyageReport) error {
	fields, err := this.reportToArgs(true)(report)
	if err != nil {
		return err
	}
	return this.update(this.updateQuery, report.Id, report.Rivers, fields)
}

func (this voyageReportStorage) Find(id int64) (VoyageReport, bool, error) {
	voyageReport, found, err := this.DoFindAndReturn(this.findQuery, readReportFromRows, id)
	if err != nil {
		return VoyageReport{}, false, err
	}
	return voyageReport.(VoyageReport), found, err
}

func (this voyageReportStorage) Remove(id int64) error {
	return this.PerformUpdates(this.removeQuery, IdMapper, id)
}

func (this voyageReportStorage) UpsertVoyageReports(reports ...VoyageReport) ([]VoyageReport, error) {
	reports_i := make([]interface{}, len(reports))
	for i := 0; i < len(reports); i++ {
		reports_i[i] = reports[i]
	}
	ids, err := this.UpdateReturningId(this.upsertQuery, func(entity interface{}) ([]interface{}, error) {
		_report := entity.(VoyageReport)
		tags, err := json.Marshal(_report.Tags)
		if err != nil {
			return []interface{}{}, err
		}
		return []interface{}{_report.Title, _report.RemoteId, _report.Source, _report.Url, _report.DatePublished,
			_report.DateModified, _report.DateOfTrip, tags, _report.Author}, nil
	}, true, reports_i...)

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
	lastDate, found, err := this.DoFindAndReturn(this.getLastIdQuery, func(rows *sql.Rows) (time.Time, error) {
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
	return this.PerformUpdates(this.upsertRiverLinkQuery, ArrayMapper, []interface{}{voyageReportId, riverId})
}

func (this voyageReportStorage) ByRiver(riverId int64, limitByGroup int) ([]VoyageReport, error) {
	result, err := this.DoFindList(this.listByRiverQuery, readReportFromRows, riverId, limitByGroup)

	if err != nil {
		return []VoyageReport{}, err
	}
	return result.([]VoyageReport), nil
}

func (this voyageReportStorage) RemoveRiverLink(id int64, tx interface{}) error {
	log.Infof("Remove spot %d", id)
	return this.PerformUpdatesWithinTxOptionally(tx, this.deleteRiverLinkQuery, IdMapper, id)
}

func readReportFromRows(rows *sql.Rows) (VoyageReport, error) {
	report := VoyageReport{}
	dateOfTrip := pq.NullTime{}
	datePublished := pq.NullTime{}
	var tags string
	rivers := pq.Int64Array{}

	err := rows.Scan(&report.Id, &report.Title, &report.RemoteId, &report.Source, &report.Url, &datePublished,
		&report.DateModified, &dateOfTrip, &tags, &report.Author, &rivers)
	report.DateOfTrip = nullDateToPtr(dateOfTrip)
	report.DatePublished = nullDateToPtr(datePublished)
	report.Rivers = []int64(rivers)

	if tags == "" {
		report.Tags = []string{}
	} else {
		err = json.Unmarshal([]byte(tags), &report.Tags)
		if err != nil {
			return report, err
		}
	}
	return report, err
}

func (this voyageReportStorage) reportToArgs(withId bool) func(entity interface{}) ([]interface{}, error) {
	return func(entity interface{}) ([]interface{}, error) {
		_report := entity.(VoyageReport)
		tags, err := json.Marshal(_report.Tags)
		if err != nil {
			return []interface{}{}, err
		}
		params := []interface{}{
			_report.Title,
			_report.RemoteId,
			_report.Source,
			_report.Url,
			_report.DatePublished,
			_report.DateModified,
			_report.DateOfTrip,
			tags,
			_report.Author,
		}

		if withId {
			params = append([]interface{}{nullIf0(_report.Id)}, params...)
		}
		return params, nil
	}
}

func convertVoyageReports(reports *[]VoyageReport) []ILinkedEntity {
	result := make([]ILinkedEntity, len(*reports))
	for i := 0; i < len(*reports); i++ {
		result[i] = &(*reports)[i]
	}
	return result
}
