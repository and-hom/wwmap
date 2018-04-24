package main

import (
	"encoding/xml"
	"fmt"
	"github.com/kokardy/saxlike"
	"os"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"log"
	"github.com/and-hom/wwmap/data"
	"strconv"
	"github.com/and-hom/wwmap/backend/model"
)

type NodeSearchHandler struct {
	saxlike.VoidHandler

	grade      string
	name       string
	rname      string
	lat        float64
	lon        float64
	_type      string
	id	   int64

	comment    string

	cnt        int
	foundcnt   int
	Node       bool
	Found      bool

	result_buf []dao.WhiteWaterPoint
	buf_size   int
	store      func([]dao.WhiteWaterPoint)
}

func (h *NodeSearchHandler) StartElement(element xml.StartElement) {
	if element.Name.Local == "node" {
		h.cnt += 1
		h.Node = true
		h.name = ""
		h.grade = ""
		h.lat = -1
		h.lon = -1
		h._type = ""

		var err error
		h.lat, err = strconv.ParseFloat(data.Attr(element.Attr, "lat"), 64)
		if err != nil {
			log.Print("Lat not found", err)
			h.Found = false
			return
		}
		h.lon, err = strconv.ParseFloat(data.Attr(element.Attr, "lon"), 64)
		if err != nil {
			log.Print("Lat not found", err)
			h.Found = false
			return
		}
		h.id, err = strconv.ParseInt(data.Attr(element.Attr, "id"), 10, 64)
		if err != nil {
			log.Print("Id found", err)
			h.Found = false
			return
		}
	}
	if h.Node && element.Name.Local == "tag" && hasEqAttr(element.Attr, "k", "whitewater") {
		h.foundcnt += 1
		h.Found = true
		h._type = data.Attr(element.Attr, "v")
	}
	if h.Node && element.Name.Local == "tag" && hasEqAttr(element.Attr, "k", "whitewater:grade_ussr") {
		h.grade = data.Attr(element.Attr, "v")
	}
	if h.Node && element.Name.Local == "tag" && hasEqAttr(element.Attr, "k", "whitewater:rapid_name") {
		h.rname = data.Attr(element.Attr, "v")
	}
	if h.Node && element.Name.Local == "tag" && hasEqAttr(element.Attr, "k", "name") {
		h.name = data.Attr(element.Attr, "v")
	}
}

func hasEqAttr(attrs []xml.Attr, name string, value string) bool {
	for _, attr := range attrs {
		if attr.Name.Local == name && attr.Value == value {
			return true
		}
	}
	return false
}

func (h *NodeSearchHandler) EndElement(element xml.EndElement) {
	if element.Name.Local == "node" {
		if h.cnt % 100000 == 0 {
			fmt.Fprintf(os.Stderr, "objects processed %d\tobjects filtered %d\tratio %.5f%%\n", h.cnt, h.foundcnt, float64(h.foundcnt) / float64(h.cnt))
		}
		if h.Found && h.name != "" {
			category := model.SportCategory{}
			err := category.UnmarshalJSON([]byte(h.grade))
			if err != nil {
				log.Fatal(h.grade, err)
			}
			var name = h.name
			if len(name) == 0 {
				name = h.rname
			}

			point := dao.WhiteWaterPoint{
				OsmId: h.id,
				Point:geo.Point{
					Lat:h.lat,
					Lon: h.lon,
				},
				Title:name,
				Category:category,
				Type:h._type,
				Comment:h.comment,
			}
			h.result_buf = append(h.result_buf, point)

			if len(h.result_buf) >= h.buf_size {
				h.flush()
			}
		}
		h.Node = false
		h.Found = false
	}
}

func (h *NodeSearchHandler) EndDocument() {
	if len(h.result_buf) > 0 {
		h.flush()
	}
}

func (h* NodeSearchHandler) flush() {
	h.store(h.result_buf)
	h.result_buf = nil
}

