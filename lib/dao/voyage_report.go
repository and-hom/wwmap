package dao

import (
	"github.com/and-hom/wwmap/lib/model"
	"database/sql"
	"time"
)

type VoyageReportStorage struct {
	PostgresStorage
}

func (this *VoyageReportStorage) UpsertVoyageReports(reports ...model.VoyageReport) ([]model.VoyageReport, error) {
	reports_i := make([]interface{}, len(reports))
	for i := 0; i < len(reports); i++ {
		reports_i[i] = reports[i]
	}
	ids,err := this.updateReturningId("INSERT INTO voyage_report(remote_id,source,url,date_published,date_modified) " +
		"VALUES ($1, $2, $3, $4, $5) " +
		"ON CONFLICT (remote_id) DO UPDATE SET url=$3, date_modified=$5 " +
		"RETURNING id",
		func(entity interface{}) ([]interface{}, error) {
			_report := entity.(model.VoyageReport)
			return []interface{}{_report.RemoteId, _report.Source, _report.Url, _report.DatePublished, _report.DateModified}, nil
		}, reports_i...)

	if err!=nil {
		return []model.VoyageReport{}, err
	}

	result := make([]model.VoyageReport, len(reports))
	copy(result, reports)
	for i := 0; i < len(reports); i++ {
		result[i].Id = ids[i]
	}
	return result, nil
}

func (this *VoyageReportStorage) GetLastId() (interface{}, error) {
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

func (this *VoyageReportStorage) AssociateWithRiver(voyageReportId, riverId int64) error {
	return this.performUpdates("INSERT INTO voyage_report_river(voyage_report_id, river_id) VALUES($1,$2) ON CONFLICT DO NOTHING",
		func(entity interface{}) ([]interface{}, error) {
			return entity.([]interface{}), nil
		}, []interface{}{voyageReportId, riverId})
}
