package model

import "time"

type VoyageReport struct {
	RiverId       int64
	RemoteId      string
	Source        string
	Url           string
	DatePublished time.Time
	DateModified time.Time
}

type Img struct {
	WwId       int64
	Source     string
	Url        string
	PreviewUrl string
	DateTaken  time.Time
}

type WWPassport struct {
	WwId int64
	Url  string
}
