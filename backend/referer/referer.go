package referer

import (
	"github.com/patrickmn/go-cache"
	"time"
	"net/url"
	"github.com/and-hom/wwmap/lib/dao"
)

type RefererStorage interface {
	Put(url *url.URL) error
	List() ([]dao.SiteRef, error)
}

const TTL time.Duration = 31 * 24 * time.Hour
const MEMORY_CLEANUP_PERIOD time.Duration = 24 * time.Hour

func CreateDummyReferrerStorage() RefererStorage {
	return dummy{
		cache: cache.New(TTL, MEMORY_CLEANUP_PERIOD),
	}
}

type dummy struct {
	cache *cache.Cache
}

func (this dummy) Put(url *url.URL) error {
	key := url.Host
	/// Unsafe. Replace with threadsafe code (CAS)
	if url.Scheme == "http" {
		existing, found := this.cache.Get(key)
		if found && existing.(dao.SiteRef).Scheme == "https" {
			// Update insert time
			this.cache.Set(key, existing, 0)
			return nil
		}
	}

	if url.Scheme == "" {
		url.Scheme = "http"
	}
	this.cache.Set(key, dao.SiteRef{
		Scheme:  url.Scheme,
		BaseUrl: url.Scheme + "://" + url.Host,
		PageUrl: url.Scheme + "://" + url.Host + url.Path,
	}, 0)
	return nil
}

func (this dummy) List() ([]dao.SiteRef, error) {
	items := this.cache.Items()
	refs := make([]dao.SiteRef, 0, len(items))
	for _, r := range items {
		refs = append(refs, r.Object.(dao.SiteRef))
	}
	return refs, nil
}

func CreateDbReferrerStorage(postgresStorage dao.PostgresStorage) RefererStorage {
	return db{dao: dao.NewRefererPostgresDao(postgresStorage), }
}

type db struct {
	dao dao.RefererDao
}

func (this db) Put(url *url.URL) error {
	if url.Scheme == "" {
		url.Scheme = "http"
	}

	baseUrl := url.Scheme + "://" + url.Host
	return this.dao.Put(url.Host, dao.SiteRef{
		Scheme:  url.Scheme,
		BaseUrl: baseUrl,
		PageUrl: baseUrl + url.Path,
	})
}

func (this db) List() ([]dao.SiteRef, error) {
	return this.dao.List(TTL)
}