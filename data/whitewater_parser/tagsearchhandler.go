package main

import (
	"encoding/xml"
	"fmt"
	"github.com/kokardy/saxlike"
	"os"
	"github.com/and-hom/wwmap/backend/dao"
	"github.com/and-hom/wwmap/backend/geo"
	"log"
	"github.com/and-hom/wwmap/data"
	"strconv"
)

type NodeSearchHandler struct {
	saxlike.VoidHandler

	grade    string
	name     string
	rname    string
	lat      float64
	lon      float64
	_type    string

	comment string

	cnt      int
	foundcnt int
	Node     bool
	Found    bool

	result   []dao.WhiteWaterPoint
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
			fmt.Fprintf(os.Stderr, "%d\t%d\t%d%%\n", h.cnt, h.foundcnt, int(float64(h.foundcnt) * 100.0 / float64(h.cnt)))
		}
		if h.Found {
			category := dao.SportCategory{}
			err := category.UnmarshalJSON([]byte(h.grade))
			if err != nil {
				log.Fatal(h.grade, err)
			}
			var name = h.name
			if len(name) == 0 {
				name = h.rname
			}

			point := dao.WhiteWaterPoint{
				Point:geo.Point{
					Lat:h.lat,
					Lon: h.lon,
				},
				Title:name,
				Category:category,
				Type:h._type,
				Comment:h.comment,
			}
			h.result = append(h.result, point)
		}
		h.Node = false
		h.Found = false
	}
}

