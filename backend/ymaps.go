package main

import (
	"strings"
	"fmt"
	"strconv"
	"bytes"
	"math"
	"github.com/Sirupsen/logrus"
)

type Bbox struct {
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
}

func (this Bbox) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[[")
	buffer.WriteString(fmt.Sprint(this.X1))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprint(this.Y1))
	buffer.WriteString("],[")
	buffer.WriteString(fmt.Sprint(this.X2))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprint(this.Y2))
	buffer.WriteString("]]")
	return buffer.Bytes(), nil
}

func NewBbox(data string) (Bbox, error) {
	parts := strings.Split(data, ",")
	if len(parts) != 4 {
		return Bbox{}, fmt.Errorf("%s is illegal bbox representation. Length sould be equals %d", data, len(parts))
	}
	partsF := make([]float64, 4)
	for i := 0; i < 4; i++ {
		partF, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return Bbox{}, fmt.Errorf("%s is illegal bbox representation. Length sould be equals %d", data, len(parts))
		}
		partsF[i] = partF
	}

	return Bbox{
		X1:partsF[0],
		Y1:partsF[1],
		X2:partsF[2],
		Y2:partsF[3],
	}, nil
}

func (this Bbox) Center() Point {
	return Point{lat:(this.X1 + this.X2) / 2, lon:(this.Y1 + this.Y2) / 2, }

}

func (this Bbox) Contains(p Point) bool {
	return this.X1 <= p.lat && p.lat <= this.X2 && this.Y1 <= p.lon && p.lon <= this.Y2;
}

type BboxInt struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

type PointInt struct {
	x int
	y int
}

type Geometry interface {

}

type yRectangleInt struct {
	Type        GeometryType `json:"type"`
	Coordinates [][]int `json:"coordinates"`
}

func NewYRectangleInt(rectangle [][]int) Geometry {
	return yRectangleInt{
		Coordinates:rectangle,
		Type:RECTANGLE,
	}
}

type FeatureProperties struct {
	BalloonContent  string `json:"balloonContent,omitempty"`
	ClusterCaption  string `json:"clusterCaption,omitempty"`
	HintContent     string `json:"hintContent,omitempty"`
	HotspotMetaData HotspotMetaData `json:"HotspotMetaData,omitempty"`
	Id              int64 `json:"id,omitempty"`
}

type FeatureOptions struct {
	Preset string `json:"preset,omitempty"`
	Id     int64 `json:"id,omitempty"`
}

type Feature struct {
	Type       string `json:"type"`
	Id         int64 `json:"id,omitempty"`
	Geometry   Geometry `json:"geometry,omitempty"`
	Properties FeatureProperties `json:"properties,omitempty"`
	Options    FeatureOptions `json:"options,omitempty"`
}

type FeatureCollection struct {
	Features []Feature `json:"features"`
	Type     string `json:"type"`
}

func (this EventPointType) toYmapsPreset() string {
	switch this {
	case PHOTO:
		return "islands#blueVegetationIcon";
	case VIDEO:
		return "islands#blueVideoIcon";
	case POST:
		return "islands#blueBookIcon";
	}
	return "islands#blueDotIcon";
}

func routeToYmaps(route Route) []Feature {
	pointCount := len(route.Points)
	trackCount := len(route.Tracks)
	featureCount := pointCount + trackCount
	result := make([]Feature, featureCount)

	var i = 0;
	for ; i < pointCount; i++ {
		point := route.Points[i]
		result[i] = Feature{
			Id:point.Id,
			Geometry:NewGeoPoint(point.Point),
			Type:"Feature",
			Properties:FeatureProperties{
				HintContent: point.Title,
				Id: point.Id,
			},
			Options:FeatureOptions{
				Preset: point.Type.toYmapsPreset(),
				Id: point.Id,
			},
		}
	}
	for ; i < featureCount; i++ {
		track := route.Tracks[i - pointCount]
		result[i] = Feature{
			Id:track.Id,
			Geometry:NewLineString(track.Path),
			Type:"Feature",
		}
	}

	return result
}

func tracksToYmaps(tracks []Track) []Feature {
	tLen := len(tracks)

	result := make([]Feature, tLen)
	for i := 0; i < tLen; i++ {
		track := tracks[i]
		result[i] = Feature{
			Id:track.Id,
			Geometry:NewLineString(track.Path),
			Type:"Feature",
		}
	}
	return result
}

func pointsToYmaps(points []EventPoint) []Feature {
	pLength := len(points)

	result := make([]Feature, pLength)
	for i := 0; i < pLength; i++ {
		point := points[i]
		result[i] = Feature{
			Id:point.Id,
			Geometry:NewGeoPoint(point.Point),
			Type:"Feature",
			Properties:FeatureProperties{
				HintContent: point.Title,
				Id: point.Id,
			},
			Options:FeatureOptions{
				Preset: point.Type.toYmapsPreset(),
				Id: point.Id,
			},
		}
	}

	return result
}

func RoutesToYmaps(route []Route) FeatureCollection {
	var features = []Feature{}
	for i := 0; i < len(route); i++ {
		features = append(features, routeToYmaps(route[i])...)
	}
	return mkFeatureCollection(features)
}

func mkFeatureCollection(features []Feature) FeatureCollection {
	return FeatureCollection{
		Features:features,
		Type:"FeatureCollection",
	}
}

func flatten(arrs [][]Feature) []Feature {
	totalSize := 0
	for i := 0; i < len(arrs); i++ {
		totalSize += len(arrs[i])
	}
	result := make([]Feature, totalSize)
	pos := 0
	for i := 0; i < len(arrs); i++ {
		for j := 0; j < len(arrs[i]); j++ {
			result[pos] = arrs[i][j]
			pos++
		}
	}
	return result
}

type HotspotMetaData struct {
	Id               int64 `json:"id"`
	RenderedGeometry Geometry `json:"RenderedGeometry"`
}

type FeatureCollectionWrapper struct {
	FeatureCollection FeatureCollection `json:"data"`
}

func tileToCoords(x int, y int, z uint32) BboxInt {
	tileCount := 1 << (z - 1)

	minX := -180
	minY := 90
	xSize := 360 / tileCount
	ySize := 180 / tileCount

	return BboxInt{
		X1 : minX + (x - 1) * xSize,
		Y1 : minY - (y - 1) * ySize,
		X2 : minX + x * xSize,
		Y2 : minY - y * ySize,
	}
}

//func toTileCoords(z uint32, p Point) (PointInt, int, int) {
//	tileCount := 1 << z
//	logrus.Infof("z=%d tileCount=%d point=%v", z, tileCount, p)
//
//	xPart := ((p.y + 180.0) / 360.0)
//	yPart := (1.0 - math.Sin(p.x * math.Pi / 180.0))
//
//	logrus.Infof("ypart=%d", math.Sin(p.x * math.Pi / 180.0))
//
//	x := math.Mod(xPart * float64(tileCount), float64(tileCount))
//	y := math.Mod(yPart * float64(tileCount), float64(tileCount))
//
//	tilePoint := PointInt{
//		x: int(math.Mod(x, 1.0) * 256.0),
//		y: int(math.Mod(y, 1.0) * 256.0),
//	}
//	logrus.Infof("point=%v x=%d,y=%d", tilePoint, int(x), int(y))
//
//	return tilePoint, int(x), int(y)
//}


//func toTileCoords1(z uint32, p Point) (PointInt, int, int) {
//	tileCount := 1 << z
//	logrus.Infof("z=%d tileCount=%d point=%v", z, tileCount, p)
//
//	xPart := ((p.y + 180.0) / 360.0)
//
//	lat := p.x
//	if lat > 89.5 {
//		lat = 89.5
//	}
//	rlat := lat * math.Pi / 180.0
//
//	maxMapY := 15496570.739634648
//	a := 6378137.0
//	b := 6356752.3142
//	f := (a - b) / a
//	e := math.Sqrt(2.0 * f - math.Pow(f, 2.0))
//	mapY := a * math.Log(
//		math.Tan(math.Pi / 4 + rlat / 2) *
//			math.Pow(
//				(1 - e * math.Sin(rlat)) / (1 + e * math.Sin(rlat)), e / 2))
//
//	yPart := (1.0 - mapY / maxMapY) / 2
//
//	logrus.Infof("mapY=%10d", mapY / maxMapY)
//
//	x := math.Mod(xPart * float64(tileCount), float64(tileCount))
//	y := math.Mod(yPart * float64(tileCount), float64(tileCount))
//
//	tilePoint := PointInt{
//		x: int(math.Mod(x, 1.0) * 256.0),
//		y: int(math.Mod(y, 1.0) * 256.0),
//	}
//	logrus.Infof("point=%v x=%d,y=%d", tilePoint, int(x), int(y))
//
//	return tilePoint, int(x), int(y)
//}

func toTileCoords(z uint32, p Point) (PointInt, int, int) {
	ppm := pixelsPerMeter(z)

	mercatorXPixels := longitudeToMercatorMeters(p.lon) * ppm
	mercatorYPixels := (EQUATOR / 2 - latitudeToMercatorMeters(p.lat)) * ppm

	logrus.Info(latitudeToMercatorMeters(p.lat), ppm)

	x := int(mercatorXPixels / 256.0)
	y := int(mercatorYPixels / 256.0)
	tilePoint := PointInt{
		int(math.Mod(mercatorXPixels, 256.0)),
		int(math.Mod(mercatorYPixels, 256.0)),
	}

	logrus.Info(tilePoint, x, y)
	return tilePoint, x, y
}

func longitudeToMercatorMeters(lon float64) float64 {
	longitudeRad := (180.0 + lon) * math.Pi / 180.0
	return RADIUS * longitudeRad
}

func latitudeToMercatorMeters(lat float64) float64 {
	latitudeRad := normalizeLatitude(lat) * math.Pi / 180.0;
	esinLat := EXCENTRICITET * math.Sin(latitudeRad);

	// Для широты -90 получается 0, и в результате по широте выходит -Infinity
	tan_temp := math.Tan(math.Pi * 0.25 + latitudeRad * 0.5);
	pow_temp := math.Pow(math.Tan(math.Pi * 0.25 + math.Asin(esinLat) * 0.5), EXCENTRICITET);
	U := tan_temp / pow_temp;

	return RADIUS * math.Log(U);
}

func normalizeLatitude(lat float64) float64 {
	if lat > 90 - LAT_EPSILON {
		return 90 - LAT_EPSILON
	}
	if lat < LAT_EPSILON - 90 {
		return LAT_EPSILON - 90
	}
	return lat
}

func pixelsPerMeter(zoom uint32) float64 {
	globalPixelCount := 1 << (zoom + 8)
	return float64(globalPixelCount) * DIV_EQUATOR;
}

const RADIUS = 6378137.0
const EXCENTRICITET = 0.0818191908426
const EQUATOR = 2.0 * math.Pi * RADIUS
const DIV_EQUATOR = 1 / EQUATOR
const LAT_EPSILON = 1e-10
