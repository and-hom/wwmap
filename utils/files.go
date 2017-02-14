package utils

import (
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"errors"
	"text/template"
	"bytes"
	"gopkg.in/yaml.v2"
)

type StringToInterfaceMapper func(data string, filename string) (interface{}, error)

func loadSingleFile(filename string, mapper StringToInterfaceMapper) (interface{}, error) {
	log.Infof("Loading data from %s", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Warnf("Can not open %s: %v", filename, err)
		return nil, err
	}

	return mapper(string(data), filename)
}

func LoadFirstSuccess(paths []string, mapper StringToInterfaceMapper) (interface{}, error) {
	var t interface{}
	var err error

	for _, path := range paths {
		t, err = loadSingleFile(path, mapper)
		if err == nil {
			return t, nil
		} else {
			log.Warnf("Can not load template from %s", path)
		}
	}
	return nil, errors.New("No sutable template found")
}

func LoadTemplate(paths []string) (*template.Template, error) {
	t, e := LoadFirstSuccess(paths,
		func(data string, filename string) (interface{}, error) {
			t, err := template.New("config").Parse(string(data))
			if err != nil {
				log.Errorf("Template error for %s: %v", filename, err)
				return nil, err
			}
			return t, nil
		})
	return t.(*template.Template), e
}

func LoadTemplatedConfig(paths []string, vars map[string]string, config interface{}) (string, error) {
	filenamePtr, err := LoadFirstSuccess(paths,
		func(data string, filename string) (interface{}, error) {
			t, err := template.New("config").Parse(string(data))
			if err != nil {
				log.Errorf("Template error for %s: %v", filename, err)
				return nil, err
			}

			templatizedConfig := bytes.Buffer{}
			tErr := t.Execute(&templatizedConfig, vars)
			if tErr != nil {
				log.Errorf("Template error for %s: %v", filename, err)
				return nil, err
			}
			log.Infof("Configuration file contents with replaced placeholder:\n%s", templatizedConfig.String())

			uErr := yaml.Unmarshal(templatizedConfig.Bytes(), config)
			if uErr != nil {
				log.Errorf("Can not unmarshal %s: %v", filename, uErr)
				return nil, uErr
			}
			return &filename, nil
		})
	filename := *(filenamePtr.(*string))
	return filename, err
}


