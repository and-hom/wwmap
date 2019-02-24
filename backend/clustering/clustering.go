package clustering

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"gopkg.in/dc0d/tinykv.v4"
	"math"
	"reflect"
	"time"
)

type RiverClusters map[ClusterId]interface{}

func NewClusterMaker(clusterizationParams config.ClusterizationParams) ClusterMaker {
	cache := tinykv.New(time.Millisecond)
	return ClusterMaker{
		ClusterizationParams: clusterizationParams,
		cache:                &cache,
	}
}

type ClusterMaker struct {
	WhiteWaterDao        dao.WhiteWaterDao
	ImgDao               dao.ImgDao
	ClusterizationParams config.ClusterizationParams
	cache                *tinykv.KV
}

func (this *ClusterMaker) Get(river dao.RiverWithSpots, zoom int, bbox geo.Bbox) (RiverClusters, error) {
	cacheKey := fmt.Sprintf("%d-%d", river.Id, zoom)

	cl, foundInCache := (*this.cache).Get(cacheKey)
	var clusters RiverClusters

	if !foundInCache {
		log.Debugf("Clusters for river %d for zoom %d not found in cache", river.Id, zoom)
		var err error
		clusters, err = this.doCluster(river, zoom)
		if err != nil {
			return clusters, err
		}
		(*this.cache).Put(cacheKey, clusters, tinykv.ExpiresAfter(15*time.Second))
	} else {
		clusters = cl.(RiverClusters)
	}

	return this.filter(clusters, bbox), nil
}

func (this *ClusterMaker) filter(riverClusters RiverClusters, bbox geo.Bbox) RiverClusters {
	result := make(RiverClusters)
	for clusterId, obj := range riverClusters {
		switch obj.(type) {
		case dao.Spot:
			if bbox.Contains(obj.(dao.Spot).Point.Center()) {
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

func (this *ClusterMaker) doCluster(river dao.RiverWithSpots, zoom int) (RiverClusters, error) {
	result := make(RiverClusters)

	riverHasSinglePoint := len(river.Spots) == 1
	actualMinDist, actualMaxDist := this.minDistance(river.Spots)
	riverClusters := []Cluster{}
	clusteredPointsCount := 0

PointsLoop:
	for i := 0; i < len(river.Spots); i++ {
		if actualMinDist < this.barrierDistance(zoom) || actualMaxDist < this.clusteringMinDistance(zoom) {
			for j := 0; j < len(riverClusters); j++ {
				if riverClusters[j].Center.DistanceTo(river.Spots[i].Point.Center()) < this.clusteringMinDistance(zoom) {
					riverClusters[j].Points = append(riverClusters[j].Points, river.Spots[i])
					riverClusters[j].RecalculateCenter()
					clusteredPointsCount++
					continue PointsLoop
				}
			}
		}
		riverClusters = append(riverClusters, Cluster{
			Center: river.Spots[i].Point.Center(),
			Points: []dao.Spot{river.Spots[i]},
		})
	}

	clusterCount := 0
	for i := 0; i < len(riverClusters); i++ {
		if len(riverClusters[i].Points) > 1 {
			clusterCount++
		}
	}

	if float64(clusteredPointsCount)/float64(len(river.Spots)) < this.ClusterizationParams.MinCLusteredPointsRatio || clusterCount > this.ClusterizationParams.MaxClustersPerRiver {
		riverClusters = make([]Cluster, len(river.Spots))
		for i := 0; i < len(river.Spots); i++ {
			riverClusters[i] = Cluster{
				Center: river.Spots[i].Point.Center(),
				Points: []dao.Spot{river.Spots[i]},
			}
		}
	}

	for idx, cluster := range riverClusters {
		if len(cluster.Points) > 1 || riverHasSinglePoint && zoom <= this.ClusterizationParams.SinglePointClusteringMaxZoom {
			clusterId := ClusterId{
				RiverId: river.Id,
				Id:      idx,
				Title:   "Cluster",
				Single:  len(riverClusters) == 1,
			}
			result[clusterId] = cluster
		} else {
			clusterId := ClusterId{
				RiverId: river.Id,
				Id:      idx,
				Title:   "Cluster",
				Single:  len(riverClusters) == 1,
			}
			result[clusterId] = cluster.Points[0]
		}
	}

	return result, nil
}

func (this *ClusterMaker) minDistance(points []dao.Spot) (float64, float64) {
	actualMinDist := math.MaxFloat64
	actualMaxDist := 0.0

	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			dist := points[i].Point.Center().DistanceTo(points[j].Point.Center())
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
	Points []dao.Spot
}

func (this *Cluster) RecalculateCenter() {
	var latSum = float64(0)
	var lonSum = float64(0)
	for i := 0; i < len(this.Points); i++ {
		latSum += this.Points[i].Point.Center().Lat
		lonSum += this.Points[i].Point.Center().Lon
	}
	this.Center = geo.Point{
		Lat: latSum / float64(len(this.Points)),
		Lon: lonSum / float64(len(this.Points)),
	}
}
