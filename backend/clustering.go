package main

import (
	"math"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
)

// CLusterization
type ClusterMaker struct {
	BarrierDistance float64 // When some points placed closer then min(width,heigth) * BarrierDistance clusterization will be applied on this river
	MinDistance     float64 // if clusterization is started for this river it's a minimal distance from cluster center to this point
}

func groupByRiver(points []dao.WhiteWaterPointWithRiverTitle) map[int64][]dao.WhiteWaterPointWithRiverTitle {
	byRiver := make(map[int64][]dao.WhiteWaterPointWithRiverTitle)
	for i := 0; i < len(points); i++ {
		riverWayId := points[i].RiverId
		byRiver[riverWayId] = append(byRiver[riverWayId], points[i])
	}
	return byRiver
}

func (this *ClusterMaker) clusterizePoints(points []dao.WhiteWaterPointWithRiverTitle, width float64, height float64) map[ClusterId][]dao.WhiteWaterPointWithRiverTitle {
	currentClustering := CurrentClustering{
		MinDistance:math.Max(width, height) * this.MinDistance,
		BarrierDistance:math.Max(width, height) * this.BarrierDistance,
	}
	result := make(map[ClusterId][]dao.WhiteWaterPointWithRiverTitle)

	for riverId, riverPoints := range groupByRiver(points) {
		riverClusters := currentClustering.clusterForRiver(riverPoints, riverId > 0 )
		for idx, cluster := range riverClusters {
			clusterId := ClusterId{
				RiverId:riverId,
				Id: idx,
				Title: "Cluster",
			}
			result[clusterId] = cluster.Points
		}
	}
	return result
}

type CurrentClustering struct {
	BarrierDistance float64 // Real value of minimal distance to __begin__ clustering for river
	MinDistance     float64 // Real value of minimal distance to __perform__ clustering for every point of river
}

func (this *CurrentClustering) clusterForRiver(points []dao.WhiteWaterPointWithRiverTitle, shouldCluster bool) []Cluster {
	riverClusters := []Cluster{}

	var actualMinDist = math.MaxFloat64
	if shouldCluster {
		for i := 0; i < len(points); i++ {
			for j := i + 1; j < len(points); j++ {
				actualMinDist = math.Min(actualMinDist, points[i].Point.DistanceTo(points[j].Point))
			}
		}
	}

	PointsLoop:
	for i := 0; i < len(points); i++ {
		if actualMinDist < this.BarrierDistance {
			for j := 0; j < len(riverClusters); j++ {
				if riverClusters[j].Center.DistanceTo(points[i].Point) < this.MinDistance {
					riverClusters[j].Points = append(riverClusters[j].Points, points[i])
					riverClusters[j].RecalculateCenter()
					continue PointsLoop
				}
			}
		}
		riverClusters = append(riverClusters, Cluster{
			Center:points[i].Point,
			Points:[]dao.WhiteWaterPointWithRiverTitle{points[i], },
		})
	}
	return riverClusters
}

type ClusterId struct {
	RiverId int64
	Id      int
	Title   string
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
