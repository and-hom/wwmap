package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/teepark/pqinterval"
	"time"
)

type SiteRef struct {
	Scheme  string `json:"scheme"`
	BaseUrl string `json:"base"`
	PageUrl string `json:"page"`
}

type refererStorage struct {
	PostgresStorage
	putQuery    string
	listQuery   string
	removeQuery string
}

func NewRefererPostgresDao(postgresStorage PostgresStorage) RefererDao {
	return refererStorage{
		PostgresStorage: postgresStorage,
		putQuery:        queries.SqlQuery("referer", "put"),
		listQuery:       queries.SqlQuery("referer", "list"),
		removeQuery:     queries.SqlQuery("referer", "remove"),
	}
}

func (this refererStorage) Put(host string, siteRef SiteRef) error {
	return this.PerformUpdates(this.putQuery, func(entity interface{}) ([]interface{}, error) {
		params := entity.([]interface{})
		_host := params[0].(string)
		_siteRef := params[1].(SiteRef)

		return []interface{}{_host, _siteRef.Scheme, _siteRef.BaseUrl, _siteRef.PageUrl}, nil
	}, []interface{}{host, siteRef})
}

func (this refererStorage) List(ttl time.Duration) ([]SiteRef, error) {
	lst, err := this.DoFindList(this.listQuery, func(rows *sql.Rows) (SiteRef, error) {
		siteRef := SiteRef{}
		_host := ""
		err := rows.Scan(&_host, &siteRef.Scheme, &siteRef.BaseUrl, &siteRef.PageUrl)
		return siteRef, err
	}, pqinterval.Duration(ttl))
	if err != nil {
		return []SiteRef{}, err
	}
	return lst.([]SiteRef), nil
}

func (this refererStorage) RemoveOlderThen(ttl time.Duration) error {
	return this.PerformUpdates(this.removeQuery, func(ttl interface{}) ([]interface{}, error) {
		return []interface{}{ttl}, nil
	}, pqinterval.Duration(ttl))
}
