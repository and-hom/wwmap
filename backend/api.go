package main

type TrackEditorPage struct {
	Title string `json:"title"`
	Type TrackType `json:"type"`
	Description string `json:"description"`
	TrackBounds Bbox `json:"trackBounds"`
	EventPoints []EventPoint `json:"eventPoints"`
}
