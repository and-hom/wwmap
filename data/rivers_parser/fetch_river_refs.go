package main

import (
	"os"
	"github.com/kokardy/saxlike"
	"log"
	"github.com/and-hom/wwmap/backend/dao"
	//"github.com/and-hom/wwmap/backend/geo"
	"github.com/and-hom/wwmap/config"
	"fmt"
)

func load_waterways(fname string, storage  dao.Storage) (map[int64]dao.WaterWayTmp, map[int64][]int64) {
	inFile, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	wayIterHandler := &WayIterHandler{}
	waysXmlParser := saxlike.NewParser(inFile, wayIterHandler)
	waysXmlParser.Parse()

	log.Printf("%d rivers loaded.", len(wayIterHandler.WayObjIndex))

	return wayIterHandler.WayObjIndex, wayIterHandler.WayObjectsByPoint
}

func load_point_refs(fname string, storage  dao.Storage, ids []int64) {

}

// river
// drain
// rapid
//

func main() {
	configuration := config.Load("")
	storage := dao.NewPostgresStorage(configuration.DbConnString)
	fname := os.Args[1]

	idx, revIdx := load_waterways(fname, storage)

	r, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	pointHandler := &PointHandler{
		waterwayIdx : idx,
		waterwayReverseIdx : revIdx,
		flush_way:func(wayId int64, ww dao.WaterWay) {
			fmt.Printf("Flush: %d", ww.Id)
			err := storage.UpdateWaterWay(ww)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	mainParser := saxlike.NewParser(r, pointHandler)
	mainParser.Parse()
}
