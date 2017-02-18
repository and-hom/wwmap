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

	r, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	mainParser := saxlike.NewParser(r, pointHandler)
	mainParser.Parse()

	waterways := make([]dao.WaterWay, len(pointHandler.result))
	var i int = 0
	for way, points := range pointHandler.result {
		waterways[i] = dao.WaterWay{
			Title:wayIterHandler.nameIndex[way],
			Type:wayIterHandler.typeIndex[way],
			Path: points,
			Comment: r.Name(),
		}
		i += 1
	}

	storage := dao.NewPostgresStorage()
	err = storage.AddWaterWays(waterways...)
	if err!=nil {
		log.Fatal(err)
	}
}
