package config

import (
	"gopkg.in/yaml.v2"
	log "github.com/Sirupsen/logrus"
	"gitlab.com/axet/desktop/go"
	"io/ioutil"
	"text/template"
	"bytes"
	"os"
	"fmt"
	"strings"
)

type Configuration struct {
	DbConnString string `yaml:"db-connection-string"`
}

func loadConf(filename string) (Configuration, error) {
	log.Infof("Loading configuration from %s", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Warnf("Can not open %s: %v", filename, err)
		return Configuration{}, err
	}
	log.Infof("Configuration file contents:\n%s", string(data))

	t, err := template.New("config").Parse(string(data))
	if err != nil {
		log.Errorf("Template error for %s: %v", filename, err)
		return Configuration{}, err
	}

	templatizedConfig := bytes.Buffer{}

	env := make(map[string]string)
	environ := os.Environ()
	for _, s := range environ {
		parts := strings.Split(s, "=")
		env[parts[0]] = parts[1]
	}
	tErr := t.Execute(&templatizedConfig, env)
	if tErr != nil {
		log.Errorf("Template error for %s: %v", filename, err)
		return Configuration{}, err
	}

	log.Infof("Configuration file contents with replaced placeholder:\n%s", templatizedConfig.String())
	config := Configuration{}
	uErr := yaml.Unmarshal(templatizedConfig.Bytes(), &config)
	if uErr != nil {
		log.Errorf("Can not unmarshal %s: %v", filename, uErr)
		return Configuration{}, uErr
	}
	return config, nil
}

func Load(configLocationOverride string) Configuration {
	if configLocationOverride != "" {
		c0, err := loadConf(configLocationOverride)
		if err == nil {
			return c0
		}
	}

	p1 := fmt.Sprintf("%s/.wwmap/config.yaml", desktop.GetHomeFolder())
	log.Infof("Try to load config from %s", p1)
	c1, err := loadConf(p1)
	if err == nil {
		return c1
	}

	p2 := "/etc/wwmap/config.yaml"
	log.Infof("Try to load config from %s", p2)
	c2, err := loadConf(p2)
	if err == nil {
		return c2
	}

	p3 := "/etc/wwmap.yaml"
	log.Infof("Try to load config from %s", p3)
	c3, err := loadConf(p3)
	if err == nil {
		return c3
	}

	p4 := "./config.yaml"
	log.Infof("Try to load config from %s", p4)
	c4, err := loadConf(p4)
	if err == nil {
		return c4
	}

	return Configuration{}
}

