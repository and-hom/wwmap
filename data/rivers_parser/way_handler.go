package main

import (
	"encoding/xml"
	"fmt"
	"github.com/kokardy/saxlike"
	"os"
	"strconv"
	"log"
	"github.com/and-hom/wwmap/data"
)


type WaterWayTmp struct {
	Id            int64 `json:"id"`
	Title         string `json:"title"`
	Type          string `json:"type"`
	ParentId      int64 `json:"parentId"`
	Comment       string `json:"comment"`
	PathPointRefs []int64 `json:"path_point_refs"`
}

var supported_types = map[string]bool{
	//"ditch", // канава
	"stream": true, // ручей
	"river": true,
	//"drain",  // дренаж
	//"riverbank",
	//"canal",
	//"dam",  // дамба
	//"weir", // водослив
	//"lake",
	//"dock",
	//"lock_gate",
	//"drystream",
	//"boatyard",
	"oxbow": true, // старица
	"water": true,
	"waterfall": true,
	"rapids": true,
	//"ditch.",
	//"wadi", // ??
	"brook": true, // ручей
	//"derelict_canal", // заброшенный канал
	//"pond", // пруд
	//"fuel" : true, // плавучая АЗС (реально в Нарьян-Маре есть)
	//"hazard", // опасность. всего один объект
	//"abandoned", // заброшенное
	//"reservoir",
	"riverbank;river" : true,
	//"riverside",
	//"tidal",

}

type WayIterHandler struct {
	saxlike.VoidHandler

	WayObjIndex       map[int64]WaterWayTmp
	WayObjectsByPoint map[int64][]int64

	way               bool
	found             bool

	cnt               int
	foundcnt          int

	currentWay        WaterWayTmp
	fileName          string
}

func (h *WayIterHandler) StartDocument() {
	h.re_init()
	h.way = false
	h.found = false
}

func (h *WayIterHandler) re_init() {
	h.WayObjIndex = make(map[int64]WaterWayTmp)
	h.WayObjectsByPoint = make(map[int64][]int64)
}

func (h *WayIterHandler) StartElement(element xml.StartElement) {
	if element.Name.Local == "way" {
		h.cnt += 1
		h.way = true
		var err error
		wayId, err := strconv.ParseInt(data.Attr(element.Attr, "id"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		h.currentWay = WaterWayTmp{Id:wayId, Comment: h.fileName}
	}
	if h.way && element.Name.Local == "tag" && data.Attr(element.Attr, "k") == "name" {
		h.currentWay.Title = data.Attr(element.Attr, "v")
	}
	if h.way && element.Name.Local == "tag" && data.Attr(element.Attr, "k") == "waterway" {
		h.currentWay.Type = data.Attr(element.Attr, "v")
		h.foundcnt += 1
		h.found = true
	}
	if h.way && element.Name.Local == "nd" {
		h.way = true
		refIdInt, err := strconv.ParseInt(data.Attr(element.Attr, "ref"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		h.currentWay.PathPointRefs = append(h.currentWay.PathPointRefs, refIdInt)
	}
}

func (h *WayIterHandler) EndElement(element xml.EndElement) {
	if h.found && h.currentWay.Title != "" && element.Name.Local == "way" {
		_, sutable_type := supported_types[h.currentWay.Type]
		if !sutable_type {
			return
		}
		if h.cnt % 1000 == 0 {
			fmt.Fprintf(os.Stderr, "%d ways processed\t%d waterways found\n", h.cnt, h.foundcnt)
		}
		h.found = false
		h.way = false

		h.WayObjIndex[h.currentWay.Id] = h.currentWay
		for _, refId := range h.currentWay.PathPointRefs {
			arr, ref_found := h.WayObjectsByPoint[refId]
			if ref_found {
				arr = append(arr, h.currentWay.Id)
			} else {
				h.WayObjectsByPoint[refId] = []int64{h.currentWay.Id}
			}
		}
	}
}

