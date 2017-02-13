package utils

import (
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"errors"
)

type StringToInterfaceMapper func(data string, filename string) (interface{}, error)

func loadSingleTemplate(filename string, mapper StringToInterfaceMapper) (interface{}, error) {
	log.Infof("Loading data from %s", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Warnf("Can not open %s: %v", filename, err)
		return nil, err
	}

	return mapper(string(data), filename)
}

func LoadTemplate(paths []string, mapper StringToInterfaceMapper) (interface{}, error) {
	var t interface{}
	var err error

	for _, path := range paths {
		t, err = loadSingleTemplate(path, mapper)
		if err == nil {
			return t, nil
		} else {
			log.Warnf("Can not load template from %s", path)
		}
	}
	return nil, errors.New("No sutable template found")
}
