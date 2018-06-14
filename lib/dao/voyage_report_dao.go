package dao

import (
	"database/sql"
	"time"
	"github.com/lib/pq"
	"encoding/json"
)

type VoyageReportStorage struct {
	PostgresStorage
}

func (this VoyageReportStorage) UpsertVoyageReports(reports ...VoyageReport) ([]VoyageReport, error) {
	reports_i := make([]interface{}, len(reports))
	for i := 0; i < len(reports); i++ {
		reports_i[i] = reports[i]
	}
	ids, err := this.updateReturningId("INSERT INTO voyage_report(title, remote_id,source,url,date_published,date_modified,date_of_trip, tags) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) " +
		"ON CONFLICT (source, remote_id) DO UPDATE SET title=$1, url=$4, date_modified=$6, date_of_trip=$7, tags=$8 " +
		"RETURNING id",
		func(entity interface{}) ([]interface{}, error) {
			_report := entity.(VoyageReport)
			tags, err := json.Marshal(_report.Tags)
			if err!=nil {
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

func (this VoyageReportStorage) GetLastId(source string) (interface{}, error) {
	lastDate := time.Unix(0, 0)
	_, err := this.doFind("SELECT max(date_modified) FROM voyage_report WHERE source=$1", func(rows *sql.Rows) error {
		rows.Scan(&lastDate)
		return nil
	}, source)
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

func (this VoyageReportStorage) List(riverId int64, limitByGroup int) ([]VoyageReport, error) {
	dateOfTrip:= pq.NullTime{}
	result, err := this.doFindList("SELECT * FROM (" +
		"SELECT  ROW_NUMBER() OVER (PARTITION BY source ORDER BY date_of_trip DESC, date_published DESC) AS r_num, " +
		"id,title,remote_id,source,url,date_published,date_modified,date_of_trip, tags " +
		"FROM voyage_report INNER JOIN voyage_report_river ON voyage_report.id = voyage_report_river.voyage_report_id " +
		"WHERE voyage_report_river.river_id = $1) sq WHERE r_num<=$2 ORDER BY source, date_of_trip DESC, date_published DESC", func(rows *sql.Rows) (VoyageReport, error) {
		report := VoyageReport{}
		var rNum int
		var tags string
		err := rows.Scan(&rNum, &report.Id, &report.Title, &report.RemoteId, &report.Source, &report.Url, &report.DatePublished, &report.DateModified, &dateOfTrip, &tags)
		if dateOfTrip.Valid {
			report.DateOfTrip = dateOfTrip.Time
		}
		err = json.Unmarshal([]byte(tags), &report.Tags)
		if err!=nil {
			return report, err
		}
		return report, err

	}, riverId, limitByGroup)

	if err != nil {
		return []VoyageReport{}, err
	}
	return result.([]VoyageReport), nil
}
