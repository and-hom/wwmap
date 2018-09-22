package config

import (
	"gopkg.in/yaml.v2"
	"os/user"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"fmt"
	"time"
)

type ClusterizationParams struct {
	BarrierRatio                 float64 `yaml:"barrier_ratio"`
	MinDistRatio                 float64 `yaml:"min_dist_ratio"`
	SinglePointClusteringMaxZoom int `yaml:"single_point_cluster_max_zoom"`
	MaxClustersPerRiver          int `yaml:"max_clusters_per_river"`
	MinCLusteredPointsRatio      float64 `yaml:"min_clustered_points_ratio"`
}

type Notifications struct {
	EmailRecipients          []string `yaml:"email_recipients"`
	EmailSender              string `yaml:"email_sender"`
	ReportingEmailSubject    string `yaml:"reporting_email_subject"`
	ImportExportEmailSubject string `yaml:"import_export_email_subject"`
}

type Api struct {
	BindTo string `yaml:"bind_to"`
}

type Content struct {
	ResourceBase string `yaml:"resource_base"`
}

type TileCache struct {
	BindTo  string `yaml:"bind_to"`
	BaseDir string `yaml:"base_dir"`
	Types   map[string][]string `yaml:"types"`
}

type WordpressSync struct {
	Login      string `yaml:"login"`
	Password   string `yaml:"password"`
	RootPageId int `yaml:"root-page-id"`
	MinDeltaBetweenRequests time.Duration `yaml:"min-delta-between-requests"`
}

type ImgStorageParams struct {
	Dir     string `yaml:"dir"`
	UrlBase string `yaml:"url-base"`
}

type ImgStorage struct {
	Full    ImgStorageParams `yaml:"full"`
	Preview ImgStorageParams `yaml:"preview"`
}

type Configuration struct {
	DbConnString         string `yaml:"db-connection-string"`
	ClusterizationParams ClusterizationParams `yaml:"clusterization"`
	Notifications        Notifications `yaml:"notifications"`
	Api                  Api `yaml:"api"`
	Content              Content `yaml:"content"`
	TileCache            TileCache `yaml:"tile-cache"`
	Sync                 WordpressSync `yaml:"sync"`
	ImgStorage           ImgStorage `yaml:"img-storage"`
	LogLevel	     string `yaml:"log-level"`
}

func (this *Configuration) ChangeLogLevel() {
	if this.LogLevel=="" {
		return
	}
	level, err := log.ParseLevel(this.LogLevel)
	if err!=nil {
		log.Fatalf("Can not parse log level %s: %v", this.LogLevel, err)
	}
	log.SetLevel(level)
}

func loadConf(filename string) (Configuration, error) {
	log.Infof("Loading configuration from %s", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Warnf("Can not open %s: %v", filename, err)
		return Configuration{}, err
	}

	config := Configuration{}
	uErr := yaml.Unmarshal(data, &config)
	if uErr != nil {
		log.Errorf("Can not unmarshal %s: %v.\nConfiguration file contents are: %s", filename, uErr, string(data))
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

	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	p1 := fmt.Sprintf("%s/.wwmap/config.yaml", currentUser.HomeDir)
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

	p5 := "../config.yaml"
	log.Infof("Try to load config from %s", p4)
	c5, err := loadConf(p5)
	if err == nil {
		return c5
	}

	return Configuration{}
}

