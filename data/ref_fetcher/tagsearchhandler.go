package main

import (
	"encoding/xml"
	"fmt"
	"bytes"
	"github.com/kokardy/saxlike"
	"os"
	"strconv"
	"log"
	"github.com/and-hom/wwmap/backend/geo"
	"github.com/and-hom/wwmap/data"
)

type PointHandler struct {
	saxlike.VoidHandler
	point_ids_by_way    map[int64][]int64
	way_ids_by_point_id map[int64][]int64
	points_by_way       map[int64][]geo.Point
	found_count_by_way  map[int64]int

	cnt                 int
	foundcnt            int
	Node                bool
	Found               bool
	Buffer              bytes.Buffer

	flush_way           func(int64, []geo.Point)
}

func (h *PointHandler) StartDocument() {
	h.points_by_way = make(map[int64]([]geo.Point))
	h.found_count_by_way = make(map[int64]int)
	//for wayId, refs := range h.point_ids_by_way {
	//	h.points_by_way[wayId] = make([]geo.Point, len(refs))
	//}
	for wayId, _ := range h.point_ids_by_way {
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
		ways, found := h.way_ids_by_point_id[id]
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
			points, _ := h.point_ids_by_way[wayId]

			pointsForWay, found := h.points_by_way[wayId]
			if !found {
				h.points_by_way[wayId] = make([]geo.Point, len(points))
			}

			for idx, p := range points {
				if p == id {
					pointsForWay[idx] = coords

					h.found_count_by_way[wayId] += 1
					if h.found_count_by_way[wayId] == len(points) {
						h.flush(wayId)
					}
				}
			}
		}
	}
}

func (h *PointHandler) EndDocument() {
	if len(h.points_by_way) > 0 {
		for k,_ := range h.points_by_way {
			h.flush(k)
		}
	}
}

func (h *PointHandler) flush(wayId int64) {
	points, _ := h.points_by_way[wayId]
	h.flush_way(wayId, points)
	delete(h.points_by_way, wayId)
}

