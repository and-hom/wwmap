package main

import (
	"github.com/kokardy/saxlike"
	"os"
	"github.com/and-hom/wwmap/backend/dao"
	"log"
)

func main() {

	r, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	handler := &NodeSearchHandler{
		comment:r.Name(),
	}
	parser := saxlike.NewParser(r, handler)
	parser.SetHTMLMode()
	parser.Parse()

	log.Print(handler.result)

	storage := dao.NewPostgresStorage()
	err = storage.AddWhiteWaterPoints(handler.result...)
	if err != nil {
		log.Fatal(err)
	}
}
