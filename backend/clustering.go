package main

import (
	"math"
	"github.com/and-hom/wwmap/backend/dao"
	"github.com/and-hom/wwmap/backend/geo"
)

const CLUSTERING_DISTANCE_KOEFF float64 = float64(0.025)

func clusterWaterWay(riverPoints []dao.WhiteWaterPoint, minDistance float64) []Cluster {
	riverClusters := []Cluster{}
	PointsLoop:
	for i := 0; i < len(riverPoints); i++ {
		for j := 0; j < len(riverClusters); j++ {
			if riverClusters[j].Center.DistanceTo(riverPoints[i].Point) < minDistance {
				riverClusters[j].Points = append(riverClusters[j].Points, riverPoints[i])
				continue PointsLoop
			}
		}
		riverClusters = append(riverClusters, Cluster{
			Center:riverPoints[i].Point,
			Points:[]dao.WhiteWaterPoint{riverPoints[i], },
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

func clusterizePoints(points []dao.WhiteWaterPoint, width float64, height float64) map[ClusterId][]dao.WhiteWaterPoint {
	minDistance := math.Max(width * CLUSTERING_DISTANCE_KOEFF, height * CLUSTERING_DISTANCE_KOEFF)
	result := make(map[ClusterId][]dao.WhiteWaterPoint)

	for waterWayId, riverPoints := range groupByRiver(points) {
		riverClusters := clusterWaterWay(riverPoints, minDistance)
		for idx, cluster := range riverClusters {
			clusterId := ClusterId{
				WaterWayId:waterWayId,
				Id: idx,
				Title: "Cluter",
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
