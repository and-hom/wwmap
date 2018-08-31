package clustering

import (
	"github.com/and-hom/wwmap/lib/geo"
	"sync"
	"github.com/and-hom/wwmap/lib/dao"
	"math"
	log "github.com/Sirupsen/logrus"
	"reflect"
	"time"
	"github.com/and-hom/wwmap/lib/config"
)

const TTL_SEC = 600

const PREVIEWS_COUNT int = 20

type RiverClusters map[ClusterId]interface{}

type CacheItem struct {
	clusters  RiverClusters
	createdAt time.Time
}

func NewClusterMaker(whiteWaterDao dao.WhiteWaterDao, imgDao dao.ImgDao, clusterizationParams config.ClusterizationParams) ClusterMaker {
	return ClusterMaker{
		WhiteWaterDao:whiteWaterDao,
		ImgDao:imgDao,
		ClusterizationParams:clusterizationParams,
	}
}

type ClusterMaker struct {
	cache                sync.Map
	mutex                sync.Mutex
	WhiteWaterDao        dao.WhiteWaterDao
	ImgDao               dao.ImgDao
	ClusterizationParams config.ClusterizationParams
}

func (this *ClusterMaker) Get(riverId int64, zoom int, bbox geo.Bbox) (RiverClusters, error) {
	byRiv, _ := this.cache.LoadOrStore(riverId, sync.Map{})
	byRiver := byRiv.(sync.Map)

	rivCl, foundInCache := byRiver.Load(zoom)
	if !foundInCache || rivCl.(CacheItem).createdAt.Add(TTL_SEC * time.Second).Before(time.Now()) {
		log.Debugf("Clusters for river %d for zoom %d not found in cache", riverId, zoom)
		this.mutex.Lock()
		rivCl, foundInCache = byRiver.Load(zoom)
		if !foundInCache {
			c, err := this.doCluster(riverId, zoom)
			if err != nil {
				return c, err
			}
			rivCl = CacheItem{
				clusters:c,
				createdAt:time.Now(),
			}
			byRiver.Store(zoom, rivCl)
		}
		this.mutex.Unlock()
	}

	riverClusters := rivCl.(CacheItem).clusters
	return this.filter(riverClusters, bbox), nil
}

func (this *ClusterMaker) filter(riverClusters RiverClusters, bbox geo.Bbox) RiverClusters {
	result := make(RiverClusters)
	for clusterId, obj := range riverClusters {
		switch obj.(type) {
		case dao.WhiteWaterPointWithRiverTitle:
			if bbox.Contains(obj.(dao.WhiteWaterPointWithRiverTitle).Point) {
				result[clusterId] = obj
			}
		case Cluster:
			if bbox.Contains(obj.(Cluster).Center) {
				result[clusterId] = obj
			}
		default:
			log.Fatalf("Unknown type: %v. Should be point or cluster", reflect.TypeOf(obj))
		}

	}
	return result
}

func (this *ClusterMaker) doCluster(riverId int64, zoom int) (RiverClusters, error) {
	result := make(RiverClusters)
	points, err := this.WhiteWaterDao.ListByRiver(riverId)
	if err != nil {
		return result, err
	}
	for i := 0; i < len(points); i++ {
		imgs, err := this.ImgDao.List(points[i].Id, PREVIEWS_COUNT, dao.IMAGE_TYPE_IMAGE, true)
		if err != nil {
			log.Warnf("Can not read whitewater point images for point %d: %s", points[i].Id, err.Error())
			continue
		}
		points[i].Images = imgs
	}

	riverHasSinglePoint := len(points) == 1
	actualMinDist, actualMaxDist := this.minDistance(points)
	riverClusters := []Cluster{}
	clusteredPointsCount := 0

	PointsLoop:
	for i := 0; i < len(points); i++ {
		if actualMinDist < this.barrierDistance(zoom) || actualMaxDist < this.clusteringMinDistance(zoom) {
			for j := 0; j < len(riverClusters); j++ {
				if riverClusters[j].Center.DistanceTo(points[i].Point) < this.clusteringMinDistance(zoom) {
					riverClusters[j].Points = append(riverClusters[j].Points, points[i])
					riverClusters[j].RecalculateCenter()
					clusteredPointsCount++
					continue PointsLoop
				}
			}
		}
		riverClusters = append(riverClusters, Cluster{
			Center:points[i].Point,
			Points:[]dao.WhiteWaterPointWithRiverTitle{points[i], },
		})
	}

	clusterCount := 0
	for i := 0; i < len(riverClusters); i++ {
		if len(riverClusters[i].Points) > 1 {
			clusterCount++
		}
	}

	if float64(clusteredPointsCount) / float64(len(points)) < this.ClusterizationParams.MinCLusteredPointsRatio || clusterCount > this.ClusterizationParams.MaxClustersPerRiver {
		riverClusters = make([]Cluster, len(points))
		for i := 0; i < len(points); i++ {
			riverClusters[i] = Cluster{
				Center:points[i].Point,
				Points:[]dao.WhiteWaterPointWithRiverTitle{points[i], },
			}
		}
	}

	for idx, cluster := range riverClusters {
		if len(cluster.Points) > 1 || riverHasSinglePoint && zoom <= this.ClusterizationParams.SinglePointClusteringMaxZoom {
			clusterId := ClusterId{
				RiverId:riverId,
				Id: idx,
				Title: "Cluster",
				Single: len(riverClusters) == 1,
			}
			result[clusterId] = cluster
		} else {
			clusterId := ClusterId{
				RiverId:riverId,
				Id: idx,
				Title: "Cluster",
				Single: len(riverClusters) == 1,
			}
			result[clusterId] = cluster.Points[0]
		}
	}

	return result, nil
}

func (this *ClusterMaker) minDistance(points []dao.WhiteWaterPointWithRiverTitle) (float64, float64) {
	actualMinDist := math.MaxFloat64
	actualMaxDist := 0.0

	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			dist := points[i].Point.DistanceTo(points[j].Point)
			if dist > 0 {
				actualMinDist = math.Min(actualMinDist, dist)
				actualMaxDist = math.Max(actualMaxDist, dist)
			}
		}
	}

	return actualMinDist, actualMaxDist
}

func (this *ClusterMaker) barrierDistance(zoom int) float64 {
	return this.ClusterizationParams.BarrierRatio * math.Pow(2.0, -float64(zoom))
}

func (this *ClusterMaker) clusteringMinDistance(zoom int) float64 {
	return this.ClusterizationParams.MinDistRatio * math.Pow(2.0, -float64(zoom))
}


type ClusterId struct {
	RiverId int64
	Id      int
	Title   string
	Single  bool // this is single cluster per river
}

type Cluster struct {
	Center geo.Point
	Points []dao.WhiteWaterPointWithRiverTitle
}

func (this *Cluster) RecalculateCenter() {
	var latSum = float64(0)
	var lonSum = float64(0)
	for i := 0; i < len(this.Points); i++ {
		latSum += this.Points[i].Point.Lat
		lonSum += this.Points[i].Point.Lon
	}
	this.Center = geo.Point{
		Lat: latSum / float64(len(this.Points)),
		Lon: lonSum / float64(len(this.Points)),
	}
}