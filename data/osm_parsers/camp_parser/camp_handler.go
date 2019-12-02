package main

import (
	"encoding/xml"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/data"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/kokardy/saxlike"
	"strconv"
)

type CampHandler struct {
	saxlike.VoidHandler

	foundcnt int
	cnt      int

	id          int64
	title       string
	description string
	point       geo.Point
	tents       bool
	found       bool

	onCamp func(dao.Camp)
}

func (this *CampHandler) StartDocument() {
	this.tents = true
}

func (this *CampHandler) StartElement(element xml.StartElement) {
	var err error

	if element.Name.Local == "node" {
		if this.cnt%100000 == 0 {
			log.Debugf("%d nodes processed. %d found\n", this.cnt, this.foundcnt)
		}
		this.cnt += 1

		this.id, err = strconv.ParseInt(data.Attr(element.Attr, "id"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		lat, err := strconv.ParseFloat(data.Attr(element.Attr, "lat"), 64)
		if err != nil {
			log.Fatal(err)
		}
		lon, err := strconv.ParseFloat(data.Attr(element.Attr, "lon"), 64)
		if err != nil {
			log.Fatal(err)
		}

		this.point = geo.Point{
			Lat: lat,
			Lon: lon,
		}
	}

	if element.Name.Local == "tag" &&
		data.Attr(element.Attr, "k") == "tourism" &&
		data.Attr(element.Attr, "v") == "camp_site" {
		this.found = true
	}

	if element.Name.Local == "tag" &&
		data.Attr(element.Attr, "k") == "name" {
		this.title = data.Attr(element.Attr, "v")
	}

	if element.Name.Local == "tag" &&
		data.Attr(element.Attr, "k") == "description" {
		this.description = data.Attr(element.Attr, "v")
	}

	if element.Name.Local == "tag" &&
		data.Attr(element.Attr, "k") == "tents" &&
		data.Attr(element.Attr, "v") == "no" {
		this.tents = false
	}
}

func (this *CampHandler) EndElement(element xml.EndElement) {
	if element.Name.Local == "node" && this.found {
		if this.tents {
			this.onCamp(dao.Camp{
				IdTitle: dao.IdTitle{
					Title: this.title,
				},
				OsmId:       this.id,
				Point:       this.point,
				Description: this.description,
			})
			this.foundcnt++
		}

		this.title = ""
		this.description = ""
		this.id = -1
		this.tents = true
		this.found = false
	}
}
