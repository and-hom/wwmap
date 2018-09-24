package util

import "time"

type RateLimit interface {
	WaitIfNecessary()
}

func NewRateLimit(limit time.Duration) RateLimit {
	return &rateLimit{
		last:ZeroDateUTC(),
		limit:limit,
	}
}

type rateLimit struct {
	last  time.Time
	limit time.Duration
}

func (this *rateLimit) waitUntilNextRequest() {
	now := time.Now()
	lastRequestTs := this.last

	delta := now.Sub(lastRequestTs)
	if (delta < this.limit) {
		time.Sleep(this.limit - delta)
	}
	this.last = now
}