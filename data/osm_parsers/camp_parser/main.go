package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/kokardy/saxlike"
	"os"
	"sync"

	//"github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/config"
)

func main() {
	configuration := config.Load("")
	configuration.ChangeLogLevel()

	storage := dao.NewPostgresStorage(configuration.Db)
	campDao := dao.NewCampPostgresDao(storage)

	fname := os.Args[1]

	Loader{
		campDao: campDao,
		fname:   fname,
	}.load()
}

type Loader struct {
	campDao dao.CampDao
	fname   string
	wg      sync.WaitGroup
	channel chan dao.Camp
}

const QUEUE_SIZE = 1024

func (this Loader) load() {
	r, err := os.Open(this.fname)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	this.channel = make(chan dao.Camp, QUEUE_SIZE)

	campHandler := &CampHandler{
		onCamp: func(camp dao.Camp) {
			this.channel <- camp
		},
	}

	this.wg.Add(1)
	go this.insertCampLoop()

	mainParser := saxlike.NewParser(r, campHandler)
	err = mainParser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Parsing completed. Wait for all data inserted")

	close(this.channel)
	this.wg.Wait()
}

const INSERT_BUF_SIZE = 512

func (this *Loader) insertCampLoop() {
	defer this.wg.Done()

	buf := make([]dao.Camp, 0, INSERT_BUF_SIZE)
	for {
		ww, ok := <-this.channel
		if len(buf) >= INSERT_BUF_SIZE || !ok && len(buf) > 0 {
			log.Debug("Flush buf")
			_, err := this.campDao.InsertMultiple(buf...)
			if err != nil {
				log.Error("Can not insert (start query per row insertion): ", err)
				for _, ww := range buf {
					_, err = this.campDao.Insert(ww)
					if err != nil {
						log.Error("Can not insert row: ", ww.OsmId, err)
					}
				}

			}
			buf = buf[:0]
		}
		buf = append(buf, ww)
		if !ok {
			log.Info("Insert of camps completed!", this.wg)
			break
		}
	}
}
