package main

import (
	"github.com/and-hom/wwmap/lib/config"
	log "github.com/Sirupsen/logrus"
)

func main() {
	configuration := config.Load("")
	configuration.ChangeLogLevel()

	log.Info("Starting pdf export")
}
