package main

import (
	"os"
	"github.com/kokardy/saxlike"
	"log"
	"github.com/and-hom/wwmap/backend/dao"
)

func main() {
	inFile, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	wayIterHandler := &WayIterHandler{
	}
	waysXmlParser := saxlike.NewParser(inFile, wayIterHandler)
	waysXmlParser.Parse()

	log.Printf("%d rivers loaded. Should detect %d points", len(wayIterHandler.wayIndex), len(wayIterHandler.reverseWayIndex))

	pointHandler := &PointHandler{
		wayIndex : wayIterHandler.wayIndex,
		reverseWayIndex: wayIterHandler.reverseWayIndex,
	}

	//r, err := os.Open(os.Args[1])
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer r.Close()
	r := os.Stdin

	mainParser := saxlike.NewParser(r, pointHandler)
	mainParser.Parse()

	log.Println("Create entries")
	waterways := make([]dao.WaterWay, 0)
	for way, points := range pointHandler.result {
		var shouldSkip bool = false
		for _, p := range points {
			if p.Lat == 0 && p.Lon == 0 {
				log.Printf("Broken path for way %d %v\n", way, points)
				shouldSkip = true
				break;
			}
		}
		if shouldSkip {
			continue
		}
		waterways = append(waterways, dao.WaterWay{
			Title:wayIterHandler.nameIndex[way],
			Type:wayIterHandler.typeIndex[way],
			Path: points,
			Comment: r.Name(),
		})
	}

	log.Println("Insert found rivers")
	storage := dao.NewPostgresStorage()
	err = storage.AddWaterWays(waterways...)
	if err != nil {
		log.Fatal(err)
	}
}
