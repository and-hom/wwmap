package referer

import (
	"github.com/patrickmn/go-cache"
	"time"
	"net/url"
)

type SiteRef struct {
	Scheme  string `json:"scheme"`
	BaseUrl string `json:"base"`
	PageUrl string `json:"page"`
}

type RefererStorage interface {
	Put(url *url.URL)
	List() []SiteRef
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

func (this dummy) Put(url *url.URL) {
	key := url.Host
	/// Unsafe. Replace with threadsafe code (CAS)
	if url.Scheme == "http" {
		existing, found := this.cache.Get(key)
		if found && existing.(SiteRef).Scheme == "https" {
			// Update insert time
			this.cache.Set(key, existing, 0)
			return
		}
	}

	if url.Scheme == "" {
		url.Scheme = "http"
	}
	this.cache.Set(key, SiteRef{
		Scheme:  url.Scheme,
		BaseUrl: url.Scheme + "://" + url.Host,
		PageUrl: url.Scheme + "://" + url.Host + url.Path,
	}, 0)
}

func (this dummy) List() []SiteRef {
	items := this.cache.Items()
	refs := make([]SiteRef, 0, len(items))
	for _, r := range items {
		refs = append(refs, r.Object.(SiteRef))
	}
	return refs
}