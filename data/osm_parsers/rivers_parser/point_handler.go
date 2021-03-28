package main

import (
	"bytes"
	"encoding/xml"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/data"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/kokardy/saxlike"
	"strconv"
)

type PointHandler struct {
	saxlike.VoidHandler

	waterwayIdx        map[int64]WaterWayTmp
	waterwayReverseIdx map[int64][]int64
	points_by_way      map[int64][]geo.Point

	cnt      int
	foundcnt int
	Node     bool
	Found    bool
	Buffer   bytes.Buffer

	flush_way   func(int64, dao.WaterWay)
	crossPoints []dao.WaterWayOsmRef
}

func (this *PointHandler) StartDocument() {
	this.points_by_way = make(map[int64]([]geo.Point))
	for wayId, refs := range this.waterwayIdx {
		this.points_by_way[wayId] = make([]geo.Point, len(refs.PathPointRefs))
	}
}

func (this *PointHandler) StartElement(element xml.StartElement) {
	if element.Name.Local == "node" {
		if this.cnt%100000 == 0 {
			log.Debugf("%d nodes processed. %d found\n", this.cnt, this.foundcnt)
		}
		this.cnt += 1

		id, err := strconv.ParseInt(data.Attr(element.Attr, "id"), 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		if id == 2376163737 {
			log.Info("Point ", 2376163737)
		}

		wayIds, found := this.waterwayReverseIdx[id]
		if !found {
			return
		}
		this.foundcnt += 1

		lat, err := strconv.ParseFloat(data.Attr(element.Attr, "lat"), 64)
		if err != nil {
			log.Fatal(err)
		}
		lon, err := strconv.ParseFloat(data.Attr(element.Attr, "lon"), 64)
		if err != nil {
			log.Fatal(err)
		}

		point := geo.Point{
			Lat: lat,
			Lon: lon,
		}

		for _, wayId := range wayIds {
			waterWayTmp, _ := this.waterwayIdx[wayId]

			pos := -1
			for idx, p := range waterWayTmp.PathPointRefs {
				if p == id {
					if pos >= 0 {
						log.Errorf("Point %d duplicate. wwId = %d", id, wayId)
					}
					pos = idx
					this.points_by_way[wayId][idx] = point
				}
			}

			if pos < 0 {
				log.Error("Point %d is not found for way %d", id, wayId)
				continue
			}
		}

		if len(wayIds) > 1 {
			for i := 0; i < len(wayIds); i++ {
				for j := 0; j < len(wayIds); j++ {
					if wayIds[i] != wayIds[j] {
						this.crossPoints = append(this.crossPoints, dao.WaterWayOsmRef{
							Id:         wayIds[i],
							RefId:      wayIds[j],
							CrossPoint: point,
						})
					}
				}
			}
		}
	}
}

func (this *PointHandler) EndDocument() {
	if len(this.points_by_way) > 0 {
		for k, _ := range this.points_by_way {
			this.flush(k)
		}
	}
}

func (this *PointHandler) flush(wayId int64) {
	points, _ := this.points_by_way[wayId]
	nullPointCnt := 0
	for i := 0; i < len(points); i++ {
		if points[i].Lat == 0 && points[i].Lon == 0 {
			log.Errorf("null point %d id=%d  present for way %d", i, this.waterwayIdx[wayId].PathPointRefs[i], wayId)
			nullPointCnt++
		}
	}
	// filter null points
	if nullPointCnt > 0 {
		log.Errorf("%d null points of %d present for way %d", nullPointCnt, len(points), wayId)
		pointsNew := make([]geo.Point, len(points)-nullPointCnt)
		idx := 0
		for i := 0; i < len(points); i++ {
			if !(points[i].Lat == 0 && points[i].Lon == 0) {
				pointsNew[idx] = points[i]
				idx++
			}
		}
		points = pointsNew
	}

	waterWayTmp, _ := this.waterwayIdx[wayId]
	this.flush_way(wayId, dao.WaterWay{
		OsmId:   waterWayTmp.Id,
		Title:   waterWayTmp.Title,
		Comment: waterWayTmp.Comment,
		Type:    waterWayTmp.Type,
		WaterWaySimple: dao.WaterWaySimple{
			Path: points,
		},
	})

	delete(this.points_by_way, wayId)
}
