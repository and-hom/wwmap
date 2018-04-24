package main

import (
	"encoding/xml"
	"fmt"
	"bytes"
	"github.com/kokardy/saxlike"
	"os"
	"strconv"
	"log"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/data"
	"github.com/and-hom/wwmap/lib/dao"
)

type PointHandler struct {
	saxlike.VoidHandler

	waterwayIdx        map[int64]dao.WaterWayTmp
	waterwayReverseIdx map[int64][]int64
	points_by_way      map[int64][]geo.Point
	found_count_by_way map[int64]int

	cnt                int
	foundcnt           int
	Node               bool
	Found              bool
	Buffer             bytes.Buffer

	flush_way          func(int64, dao.WaterWay)
}

func (h *PointHandler) StartDocument() {
	h.points_by_way = make(map[int64]([]geo.Point))
	h.found_count_by_way = make(map[int64]int)
	for wayId, refs := range h.waterwayIdx {
		h.points_by_way[wayId] = make([]geo.Point, len(refs.PathPointRefs))
		h.found_count_by_way[wayId] = 0
	}
}

func (h *PointHandler) StartElement(element xml.StartElement) {
	if element.Name.Local == "node" {
		if h.cnt % 100000 == 0 {
			fmt.Fprintf(os.Stderr, "%d nodes processed. %d found\n", h.cnt, h.foundcnt)
		}
		h.cnt += 1

		id, err := strconv.ParseInt(data.Attr(element.Attr, "id"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		ways, found := h.waterwayReverseIdx[id]
		if !found {
			return
		}
		h.foundcnt += 1

		lat, err := strconv.ParseFloat(data.Attr(element.Attr, "lat"), 64)
		if err != nil {
			log.Fatal(err)
		}
		lon, err := strconv.ParseFloat(data.Attr(element.Attr, "lon"), 64)
		if err != nil {
			log.Fatal(err)
		}
		coords := geo.Point{
			Lat:lat,
			Lon:lon,
		}

		for _, wayId := range ways {
			waterWayTmp, _ := h.waterwayIdx[wayId]
			pointsForWay, _ := h.points_by_way[wayId]

			for idx, p := range waterWayTmp.PathPointRefs {
				if p == id {
					pointsForWay[idx] = coords
					h.found_count_by_way[wayId] += 1
					if h.found_count_by_way[wayId] == len(waterWayTmp.PathPointRefs) {
						h.flush(wayId)
					}
				}
			}
		}
	}
}

func (h *PointHandler) EndDocument() {
	if len(h.points_by_way) > 0 {
		for k, _ := range h.points_by_way {
			h.flush(k)
		}
	}
}

func (h *PointHandler) flush(wayId int64) {
	points, _ := h.points_by_way[wayId]
	nullPointCnt := 0
	for i:=0;i<len(points);i++ {
		if points[i].Lat == 0 && points[i].Lon==0 {
			nullPointCnt++
		}
	}
	pointsNew := make([]geo.Point, len(points) - nullPointCnt)
	idx :=0
	for i:=0;i<len(points);i++ {
		if !(points[i].Lat == 0 && points[i].Lon==0) {
			pointsNew[idx] = points[i]
			idx++
		}
	}
	fmt.Printf("%d points of %d removed\n", nullPointCnt, len(points))

	waterWayTmp, _ := h.waterwayIdx[wayId]
	h.flush_way(wayId, dao.WaterWay{
		OsmId: waterWayTmp.Id,
		Title: waterWayTmp.Title,
		ParentId: waterWayTmp.ParentId,
		Comment: waterWayTmp.Comment,
		Type: waterWayTmp.Type,
		Path: pointsNew,
		Verified: false,
		Popularity: 0,
	})

	delete(h.points_by_way, wayId)
}

