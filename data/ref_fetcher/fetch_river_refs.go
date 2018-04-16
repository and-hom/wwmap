package main

import (
	"os"
	"github.com/kokardy/saxlike"
	"log"
	"github.com/and-hom/wwmap/backend/dao"
	//"github.com/and-hom/wwmap/backend/geo"
)

func load_waterways(fname string, storage  dao.Storage) {
	inFile, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	wayIterHandler := &WayIterHandler{
		flush_f: func(waterwaysTmp []dao.WaterWayTmp) {
			storage.AddTmpWaterWay(waterwaysTmp...)
		},
		comment:inFile.Name(),
		buf_size:1,
	}
	waysXmlParser := saxlike.NewParser(inFile, wayIterHandler)
	waysXmlParser.Parse()

	log.Printf("%d rivers loaded.", len(wayIterHandler.wayIndex))
}


func load_point_refs(fname string, storage  dao.Storage, ids []int64) {

}

// river
// drain
// rapid
//

func main() {
	storage := dao.NewPostgresStorage()
	fname := os.Args[1]

	load_waterways(fname, storage)

	p_ref_ids, err := storage.GetUniquePointRefIds()
	if err!=nil {
		log.Fatal(err)
	}

	load_point_refs(fname, storage, p_ref_ids)

	r, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	//
	//pointHandler := &PointHandler{
	//	point_ids_by_way : wayIterHandler.wayIndex,
	//	flush_way:func(wayId int64, points []geo.Point) {
	//		waterway := dao.WaterWay{
	//			Title:wayIterHandler.nameIndex[wayId],
	//			Type:wayIterHandler.typeIndex[wayId],
	//			Path: points,
	//			Comment: r.Name(),
	//		}
	//		storage.AddWaterWays(waterway)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//	},
	//}
	//
	//mainParser := saxlike.NewParser(r, pointHandler)
	//mainParser.Parse()
}
