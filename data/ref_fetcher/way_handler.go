package main

import (
	"encoding/xml"
	"fmt"
	"github.com/kokardy/saxlike"
	"os"
	"strconv"
	"log"
	"github.com/and-hom/wwmap/data"
	"github.com/and-hom/wwmap/backend/dao"
)

type WayIterHandler struct {
	saxlike.VoidHandler

	wayIndex  map[int64][]int64

	nameIndex map[int64]string
	typeIndex map[int64]string

	way       bool
	cnt       int
	foundcnt  int
	refIds    []int64
	wayId     int64

	found     bool
	buf_size  int
	flush_f   func([]dao.WaterWayTmp)
	comment   string
}

func (h *WayIterHandler) StartDocument() {
	h.re_init()
	h.way = false
	h.found = false
}

func (h *WayIterHandler) re_init() {
	h.wayIndex = make(map[int64]([]int64))
	h.nameIndex = make(map[int64]string)
	h.typeIndex = make(map[int64]string)
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
		h.foundcnt += 1
		h.found = true
	}
	if h.way && element.Name.Local == "nd" {
		h.way = true
		refIdInt, err := strconv.ParseInt(data.Attr(element.Attr, "ref"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		h.refIds = append(h.refIds, refIdInt)
	}
}

func (h *WayIterHandler) EndElement(element xml.EndElement) {
	if h.found && element.Name.Local == "way" {
		if h.cnt % 1000 == 0 {
			fmt.Fprintf(os.Stderr, "%d ways processed\t%d waterways found\n", h.cnt, h.foundcnt)
		}
		h.found = false
		h.way = false
		h.wayIndex[h.wayId] = h.refIds

		h.refIds = []int64{}

		if len(h.wayIndex) >= h.buf_size {
			h.flush()
		}
	}
}

func (h *WayIterHandler) EndDocument() {
	if len(h.wayIndex) >= 0 {
		h.flush()
	}
}

func (h *WayIterHandler) flush() {
	wways := make([]dao.WaterWayTmp, len(h.wayIndex))
	points := make([]dao.PointRef, 0)

	i := 0
	for id, refs := range h.wayIndex {
		wways[i] = dao.WaterWayTmp{
			Id:id,
			Title:h.nameIndex[id],
			Type:h.typeIndex[id],
			Comment:h.comment,
		}
		i += 1

		for idx,ref := range refs {
			points = append(points, dao.PointRef{
				Id:ref,
				ParentId:id,
				Idx: idx,
			})
		}


	}
	h.flush_f(wways)
	h.re_init();
}
