package main

import (
	"math"
	"github.com/and-hom/wwmap/backend/dao"
	"github.com/and-hom/wwmap/backend/geo"
)

// CLusterization
type ClusterMaker struct {
	BarrierDistance float64 // When some points placed closer then min(width,heigth) * BarrierDistance clusterization will be applied on this river
	MinDistance     float64 // if clusterization is started for this river it's a minimal distance from cluster center to this point
}


func clusterWaterWay(waterWayId int64, points []dao.WhiteWaterPoint, barrierMinDistance float64, minDistance float64) []Cluster {
	riverClusters := []Cluster{}

	var actualMinDist = math.MaxFloat64
	if waterWayId > 0 {
		for i := 0; i < len(points); i++ {
			for j := i + 1; j < len(points); j++ {
				actualMinDist = math.Min(actualMinDist, points[i].Point.DistanceTo(points[j].Point))
			}
		}
	}

	PointsLoop:
	for i := 0; i < len(points); i++ {
		if actualMinDist < barrierMinDistance {
			for j := 0; j < len(riverClusters); j++ {
				if riverClusters[j].Center.DistanceTo(points[i].Point) < minDistance {
					riverClusters[j].Points = append(riverClusters[j].Points, points[i])
					riverClusters[j].RecalculateCenter()
					continue PointsLoop
				}
			}
		}
		riverClusters = append(riverClusters, Cluster{
			Center:points[i].Point,
			Points:[]dao.WhiteWaterPoint{points[i], },
		})
	}
	return riverClusters
}

func groupByRiver(points []dao.WhiteWaterPoint) map[int64][]dao.WhiteWaterPoint {
	byRiver := make(map[int64][]dao.WhiteWaterPoint)
	for i := 0; i < len(points); i++ {
		waterWayId := points[i].WaterWayId
		byRiver[waterWayId] = append(byRiver[waterWayId], points[i])
	}
	return byRiver
}

func (this *ClusterMaker) clusterizePoints(points []dao.WhiteWaterPoint, width float64, height float64) map[ClusterId][]dao.WhiteWaterPoint {
	minDistance := math.Max(width, height) * this.MinDistance
	barrierMinDistance :=  math.Max(width, height) * this.BarrierDistance
	result := make(map[ClusterId][]dao.WhiteWaterPoint)

	for waterWayId, riverPoints := range groupByRiver(points) {
		riverClusters := clusterWaterWay(waterWayId, riverPoints, barrierMinDistance, minDistance)
		for idx, cluster := range riverClusters {
			clusterId := ClusterId{
				WaterWayId:waterWayId,
				Id: idx,
				Title: "Cluster",
			}
			result[clusterId] = cluster.Points
		}
	}
	return result
}

type ClusterId struct {
	WaterWayId int64
	Id         int
	Title      string
}

type Cluster struct {
	Center geo.Point
	Points []dao.WhiteWaterPoint
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
