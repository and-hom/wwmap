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

type WayIterHandler struct {
	saxlike.VoidHandler
	wayIndex        map[int64][]int64
	reverseWayIndex map[int64][]int64
	nameIndex       map[int64]string
	typeIndex       map[int64]string

	way             bool
	cnt             int
	refIds          []int64
	wayId           int64
}

func (h *WayIterHandler) StartDocument() {
	h.wayIndex = make(map[int64]([]int64))
	h.nameIndex = make(map[int64]string)
	h.typeIndex = make(map[int64]string)
	h.reverseWayIndex = make(map[int64]([]int64))
	h.refIds = []int64{}
}

func (h *WayIterHandler) StartElement(element xml.StartElement) {
	if element.Name.Local == "way" {
		h.cnt += 1
		h.way = true
		var err error
		h.wayId, err = strconv.ParseInt(data.Attr(element.Attr, "id"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
	}
	if h.way && element.Name.Local == "tag" && data.Attr(element.Attr, "k") == "name" {
		h.nameIndex[h.wayId] = data.Attr(element.Attr, "v")
	}
	if h.way && element.Name.Local == "tag" && data.Attr(element.Attr, "k") == "waterway" {
		h.typeIndex[h.wayId] = data.Attr(element.Attr, "v")
	}
	if h.way && element.Name.Local == "nd" {
		h.way = true
		refIdInt, err := strconv.ParseInt(data.Attr(element.Attr, "ref"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		h.refIds = append(h.refIds, refIdInt)
		found, exists := h.reverseWayIndex[refIdInt]
		if exists {
			h.reverseWayIndex[refIdInt] = append(found, h.wayId)
		} else {
			h.reverseWayIndex[refIdInt] = []int64{h.wayId}
		}
	}
}

func (h *WayIterHandler) EndElement(element xml.EndElement) {
	if element.Name.Local == "way" {
		if h.cnt % 1000 == 0 {
			fmt.Fprintf(os.Stderr, "%d rivers processed\n", h.cnt)
		}
		h.way = false
		h.wayIndex[h.wayId] = h.refIds

		h.refIds = []int64{}
	}
}

type PointHandler struct {
	saxlike.VoidHandler
	wayIndex        map[int64][]int64
	reverseWayIndex map[int64][]int64
	result          map[int64][]geo.Point

	cnt             int
	foundcnt        int
	Node            bool
	Found           bool
	Buffer          bytes.Buffer
}

func (h *PointHandler) StartDocument() {
	h.result = make(map[int64]([]geo.Point))
	for wayId, refs := range h.wayIndex {
		h.result[wayId] = make([]geo.Point, len(refs))
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
		ways, found := h.reverseWayIndex[id]
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
			points, _ := h.wayIndex[wayId]
			for idx, p := range points {
				if p == id {
					resultV, _ := h.result[wayId]
					resultV[idx] = coords
				}
			}
		}
	}
}

