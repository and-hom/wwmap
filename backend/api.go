package main

type RouteEditorPage struct {
	Id int64 `json:"id,omitempty"`
	Title string `json:"title"`
	Type TrackType `json:"type"`
	Description string `json:"description"`
	Bounds Bbox `json:"bounds"`
	Tracks []Track `json:"tracks,omitempty"`
	EventPoints []EventPoint `json:"points,omitempty"`
}
