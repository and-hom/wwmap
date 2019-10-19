package main

import (
	"encoding/xml"
	"fmt"
	"github.com/and-hom/wwmap/data"
	"github.com/kokardy/saxlike"
	"log"
	"os"
	"strconv"
)

type WaterWayTmp struct {
	Id            int64   `json:"id"`
	Title         string  `json:"title"`
	Type          string  `json:"type"`
	ParentId      int64   `json:"parentId"`
	Comment       string  `json:"comment"`
	PathPointRefs []int64 `json:"path_point_refs"`
}

var supported_types = map[string]bool{
	//"ditch", // канава
	"stream": true, // ручей
	"river":  true,
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
	"oxbow":     true, // старица
	"water":     true,
	"waterfall": true,
	"rapids":    true,
	//"ditch.",
	//"wadi", // ??
	"brook": true, // ручей
	//"derelict_canal", // заброшенный канал
	//"pond", // пруд
	//"fuel" : true, // плавучая АЗС (реально в Нарьян-Маре есть)
	//"hazard", // опасность. всего один объект
	//"abandoned", // заброшенное
	//"reservoir",
	"riverbank;river": true,
	//"riverside",
	//"tidal",

}

type WayIterHandler struct {
	saxlike.VoidHandler

	WayObjIndex       map[int64]WaterWayTmp
	WayObjectsByPoint map[int64][]int64

	way   bool
	found bool

	cnt      int
	foundcnt int

	currentWay WaterWayTmp
	fileName   string
}

func (this *WayIterHandler) StartDocument() {
	this.re_init()
	this.way = false
	this.found = false
}

func (this *WayIterHandler) re_init() {
	this.WayObjIndex = make(map[int64]WaterWayTmp)
	this.WayObjectsByPoint = make(map[int64][]int64)
}

func (this *WayIterHandler) StartElement(element xml.StartElement) {
	if element.Name.Local == "way" {
		this.cnt += 1
		this.way = true
		var err error
		wayId, err := strconv.ParseInt(data.Attr(element.Attr, "id"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		this.currentWay = WaterWayTmp{Id: wayId, Comment: this.fileName}
	}
	if this.way && element.Name.Local == "tag" && data.Attr(element.Attr, "k") == "name" {
		this.currentWay.Title = data.Attr(element.Attr, "v")
	}
	if this.way && element.Name.Local == "tag" && data.Attr(element.Attr, "k") == "waterway" {
		this.currentWay.Type = data.Attr(element.Attr, "v")
		this.foundcnt += 1
		this.found = true
	}
	if this.way && element.Name.Local == "nd" {
		this.way = true
		refIdInt, err := strconv.ParseInt(data.Attr(element.Attr, "ref"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		this.currentWay.PathPointRefs = append(this.currentWay.PathPointRefs, refIdInt)
	}
}

func (this *WayIterHandler) EndElement(element xml.EndElement) {
	if this.found && this.currentWay.Title != "" && element.Name.Local == "way" {
		_, sutable_type := supported_types[this.currentWay.Type]
		if !sutable_type {
			return
		}
		if this.cnt%1000 == 0 {
			fmt.Fprintf(os.Stderr, "%d ways processed\t%d waterways found\n", this.cnt, this.foundcnt)
		}
		this.found = false
		this.way = false

		this.WayObjIndex[this.currentWay.Id] = this.currentWay
		for _, refId := range this.currentWay.PathPointRefs {
			arr, ref_found := this.WayObjectsByPoint[refId]
			if ref_found {
				this.WayObjectsByPoint[refId] = append(arr, this.currentWay.Id)
			} else {
				this.WayObjectsByPoint[refId] = []int64{this.currentWay.Id}
			}
		}
	}
}
