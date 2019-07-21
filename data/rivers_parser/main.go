package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/kokardy/saxlike"
	"os"
	"sync"

	//"github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/config"
)

func load_waterways(fname string) (map[int64]WaterWayTmp, map[int64][]int64) {
	inFile, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()

	wayIterHandler := &WayIterHandler{
		fileName: fname,
	}
	waysXmlParser := saxlike.NewParser(inFile, wayIterHandler)
	waysXmlParser.Parse()

	log.Printf("%d rivers loaded.", len(wayIterHandler.WayObjIndex))

	return wayIterHandler.WayObjIndex, wayIterHandler.WayObjectsByPoint
}

func load_point_refs(fname string, storage dao.Storage, ids []int64) {

}

// river
// drain
// rapid
//

const QUEUE_SIZE = 1024

func main() {
	configuration := config.Load("")
	configuration.ChangeLogLevel()

	storage := dao.NewPostgresStorage(configuration.Db)
	waterWayDao := dao.NewWaterWayPostgresDao(storage)
	osmRefDao := dao.NewWaterWayOsmRefPostgresDao(storage)
	fname := os.Args[1]

	Loader{
		waterWayDao: waterWayDao,
		osmRefDao:   osmRefDao,
		fname:       fname,
	}.load()
}

type Loader struct {
	waterWayDao dao.WaterWayDao
	osmRefDao   dao.WaterWayOsmRefDao
	fname       string
	wg          sync.WaitGroup
	channel     chan dao.WaterWay
}

func (this Loader) load() {
	idx, revIdx := load_waterways(this.fname)

	r, err := os.Open(this.fname)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	this.channel = make(chan dao.WaterWay, QUEUE_SIZE)

	pointHandler := &PointHandler{
		waterwayIdx:        idx,
		waterwayReverseIdx: revIdx,
		flush_way: func(wayId int64, ww dao.WaterWay) {
			this.channel <- ww
		},
	}

	this.wg.Add(1)
	go this.insertWaterWayLoop()

	mainParser := saxlike.NewParser(r, pointHandler)
	err = mainParser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Parsing completed. Wait for all data inserted")

	close(this.channel)
	this.wg.Wait()

	log.Info("Insert OSM refs")
	err = this.osmRefDao.Insert(pointHandler.crossPoints...)
	if err != nil {
		log.Fatal(err)
	}
}

const INSERT_BUF_SIZE = 512

func (this *Loader) insertWaterWayLoop() {
	defer this.wg.Done()

	buf := make([]dao.WaterWay, 0, INSERT_BUF_SIZE)
	for {
		ww, ok := <-this.channel
		if len(buf) >= INSERT_BUF_SIZE || !ok && len(buf) > 0 {
			log.Debug("Flush buf")
			err := this.waterWayDao.AddWaterWays(buf...)
			if err != nil {
				log.Fatal("Can not insert: ", err)
				return
			}
			buf = buf[:0]
		}
		buf = append(buf, ww)
		if !ok {
			log.Info("Insert of waterways completed!", this.wg)
			break
		}
	}
}
