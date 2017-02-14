package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/backend/dao"
	"github.com/and-hom/wwmap/utils"
	"time"
	"os"
	"fmt"
	"text/template"
)

const PORTION_SIZE int = 100
const SLEEP_IF_NOTHING_FOUND time.Duration = 10 * time.Second
const OUTPUT_DIR string = "/tmp"

func main() {
	log.Infof("Starting wwmap static content exporter")

	log.Infof("Compile template")
	tmplObj, err := utils.LoadTemplate([]string{
		"/etc/wwmap/export-template",
		"template.htm",
	}, func(data string, filename string) (interface{}, error) {
		t, err := template.New("config").Parse(string(data))
		if err != nil {
			log.Errorf("Template error for %s: %v", filename, err)
			return nil, err
		}
		return t, nil
	})
	if err != nil {
		log.Errorf("Can not parse template")
	}

	tmpl := tmplObj.(*template.Template)

	log.Infof("Connect to database")
	storage := dao.NewPostgresStorage()

	log.Infof("Start processing")
	for {
		routes, err := storage.ListUnExportedRoutes(PORTION_SIZE);
		if err != nil {
			log.Errorf("Can not get unexported routes: %s", err.Error())
			continue;
		}
		if len(routes) == 0 {
			log.Infof("Nothing found. Sleep %v", SLEEP_IF_NOTHING_FOUND)
			time.Sleep(SLEEP_IF_NOTHING_FOUND)
			continue
		}
		for _, route := range routes {
			path := fmt.Sprintf("%s/%d", OUTPUT_DIR, route.Id)
			if route.Publish {
				var success bool = true
				outFile, err := os.Create(path)
				if err != nil {
					log.Errorf("Can not open results file %s for route %d: %s", path, route.Id, err.Error())
					continue
				}
				err = tmpl.Execute(outFile, route)
				if err != nil {
					success = false
					log.Errorf("Can not process template for route %d: %s", route.Id, err.Error())
				}

				err = outFile.Close()
				if err != nil {
					log.Errorf("Can not write results to file %s for route %d: %s", path, route.Id, err.Error())
					continue
				}

				if success {
					err := storage.MarkRouteExported(route.Id)
					if err != nil {
						log.Errorf("Can not mark route %d exported: %s", route.Id, err.Error())
					}
					log.Infof("Route %d exported", route.Id)
				} else {
					log.Warnf("Template processing for route %d was not successfull. Remove results.")
					err := os.Remove(path)
					if err != nil {
						log.Errorf("Can not remove to file %s for route %d: %s", path, route.Id, err.Error())
					}
				}
			} else {
				err := os.Remove(path)
				if err != nil  && !os.IsNotExist(err) {
					log.Errorf("Can not remove file %s for unpublished route %d: %s", path, route.Id, err.Error())
				} else if err == nil {
					log.Infof("Removed file %s for unpublished report %d", path, route.Id)
					err := storage.MarkRouteExported(route.Id)
					if err != nil {
						log.Errorf("Can not mark route %d exported (unpublished): %s", route.Id, err.Error())
					}
				}
			}
		}
	}
}

