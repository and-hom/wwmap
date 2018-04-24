package main

import . "github.com/and-hom/wwmap/lib/dao"
import (
	. "github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/backend/model"
)

type RouteEditorPage struct {
	Id          int64 `json:"id,omitempty"`
	Title       string `json:"title"`
	Type        TrackType `json:"type"`
	Description string `json:"description"`
	Bounds      Bbox `json:"bounds"`
	Tracks      []Track `json:"tracks,omitempty"`
	EventPoints []EventPoint `json:"points,omitempty"`
	Category    model.SportCategory `json:"category"`
}
