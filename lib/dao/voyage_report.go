package dao

import (
	"github.com/and-hom/wwmap/lib/model"
	"database/sql"
	"time"
)

type VoyageReportStorage struct {
	PostgresStorage
}

func (this VoyageReportStorage) UpsertVoyageReports(reports ...model.VoyageReport) ([]model.VoyageReport, error) {
	reports_i := make([]interface{}, len(reports))
	for i := 0; i < len(reports); i++ {
		reports_i[i] = reports[i]
	}
	ids, err := this.updateReturningId("INSERT INTO voyage_report(title, remote_id,source,url,date_published,date_modified) " +
		"VALUES ($1, $2, $3, $4, $5, $6) " +
		"ON CONFLICT (remote_id) DO UPDATE SET title=$1, url=$4, date_modified=$6 " +
		"RETURNING id",
		func(entity interface{}) ([]interface{}, error) {
			_report := entity.(model.VoyageReport)
			return []interface{}{_report.Title, _report.RemoteId, _report.Source, _report.Url, _report.DatePublished, _report.DateModified}, nil
		}, reports_i...)

	if err != nil {
		return []model.VoyageReport{}, err
	}

	result := make([]model.VoyageReport, len(reports))
	copy(result, reports)
	for i := 0; i < len(reports); i++ {
		result[i].Id = ids[i]
	}
	return result, nil
}

func (this VoyageReportStorage) GetLastId() (interface{}, error) {
	lastDate := time.Unix(0, 0)
	_, err := this.doFind("SELECT max(date_modified) FROM voyage_report", func(rows *sql.Rows) error {
		rows.Scan(&lastDate)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lastDate, nil
}

func (this VoyageReportStorage) AssociateWithRiver(voyageReportId, riverId int64) error {
	return this.performUpdates("INSERT INTO voyage_report_river(voyage_report_id, river_id) VALUES($1,$2) ON CONFLICT DO NOTHING",
		func(entity interface{}) ([]interface{}, error) {
			return entity.([]interface{}), nil
		}, []interface{}{voyageReportId, riverId})
}

func (this VoyageReportStorage) List(riverId int64) ([]model.VoyageReport, error) {
	result, err :=  this.doFindList("SELECT id,title,remote_id,source,url,date_published,date_modified " +
		"FROM voyage_report INNER JOIN voyage_report_river ON voyage_report.id = voyage_report_river.voyage_report_id " +
		"WHERE voyage_report_river.river_id = $1", func(rows *sql.Rows) (model.VoyageReport, error) {
		report := model.VoyageReport{}
		err := rows.Scan(&report.Id, &report.Title, &report.RemoteId, &report.Source, &report.Url, &report.DatePublished, &report.DateModified)
		return report, err

	}, riverId)

	if err!=nil {
		return []model.VoyageReport{}, err
	}
	return result.([]model.VoyageReport), nil
}
