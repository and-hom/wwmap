package geoparser

import (
	"io"
	"encoding/xml"
	"io/ioutil"
	"strings"
	"fmt"
	"strconv"
	. "github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/geo"
	"time"
	"github.com/Sirupsen/logrus"
)

type Kml struct {
	XMLName    xml.Name `xml:"kml"`
	Namespace  string   `xml:"xmlns,attr"`
	Document   Document
	Placemarks []Placemark `xml:"Placemark"`
	Folders    []Folder    `xml:"Folder"`
}
type Document struct {
	XMLName     xml.Name `xml:"Document"`
	Id          string   `xml:"id,attr,omitempty"`
	Name        string   `xml:"name,omitempty"`
	Visibility  int      `xml:"visibility,omitempty"`
	Open        int      `xml:"open,omitempty"`
	Address     string   `xml:"address,omitempty"`
	PhoneNumber string   `xml:"phoneNumber,omitempty"`
	Description string   `xml:"description,omitempty"`
	Placemarks  []Placemark `xml:"Placemark"`
	Folders     []Folder `xml:"Folder"`
	DocStyle    []Style `xml:"Style"`
}
type Folder struct {
	XMLName     xml.Name    `xml:"Folder"`
	Id          string      `xml:"id,attr,omitempty"`
	Name        string      `xml:"name,omitempty"`
	Visibility  int         `xml:"visibility,omitempty"`
	Open        int         `xml:"open,omitempty"`
	Address     string      `xml:"address,omitempty"`
	PhoneNumber string      `xml:"phoneNumber,omitempty"`
	Description string      `xml:"description,omitempty"`
	Styles      []Style     `xml:"Style"`
	Placemarks  []Placemark `xml:"Placemark"`
	Folders     []Folder    `xml:"Folder"`
}
type Style struct {
	XMLName xml.Name `xml:"Style"`
	Id      string   `xml:"id,attr,omitempty"`
	Icon    IconStyle
	Label   LabelStyle
}
type IconStyle struct {
	XMLName xml.Name `xml:"IconStyle"`
	Scale   string   `xml:"scale,omitempty"`
	Heading string   `xml:"heading,omitempty"`
	Href    string   `xml:"Icon>href,omitempty"`
}

type LabelStyle struct {
	XMLName xml.Name `xml:"LabelStyle"`
	Scale   string   `xml:"scale,omitempty"`
	Color   string   `xml:"color,omitempty"`
}
type Placemark struct {
	XMLName     xml.Name `xml:"Placemark"`
	Id          string   `xml:"id,attr,omitempty"`
	Name        string   `xml:"name,omitempty"`
	Description string   `xml:"description,omitempty"`
	StyleUrl    string   `xml:"styleUrl,omitempty"`
	Point       string   `xml:"Point>coordinates"`
	LineString  string   `xml:"LineString>coordinates"`
	Extended    ExtendedData
}
type ExtendedData struct {
	XMLName xml.Name `xml:"ExtendedData"`
	Datas   []Data   `xml:"Data"`
}
type Data struct {
	XMLName xml.Name `xml:"Data"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value"`
}

type KmlParser struct {
	kml_data Kml
}

type Point3 struct {
	Lat float64
	Lon float64
	Alt float64
}

func (this Point3) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d,%d,%d", this.Lon, this.Lat, this.Alt)), nil
}

func (this *Point3) UnmarshalJSON(data []byte) error {
	dataStr := string(data)
	parts := strings.Split(dataStr, ",")
	if len(parts) != 3 {
		return fmt.Errorf("Invalid KML point format: %s", dataStr)
	}
	var err error
	this.Lat, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return err
	}
	this.Lon, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return err
	}
	this.Alt, err = strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
		return err
	}
	return nil
}

func (this *Point3) toPointSwappingLatLon() Point {
	return Point{
		Lat : this.Lon,
		Lon : this.Lat,
	}
}

func InitKmlParser(reader io.Reader) (*KmlParser, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var kml_data Kml
	err = xml.Unmarshal(bytes, &kml_data)
	if err != nil {
		return nil, err
	}
	parser := KmlParser{
		kml_data:kml_data,
	}
	return &parser, nil
}

func (this KmlParser) GetTracksAndPoints() ([]Track, []EventPoint, error) {
	tracks := make([]Track, 0)
	points := make([]EventPoint, 0)

	for _, folder := range this.kml_data.Document.Folders {
		for _, placemark := range folder.Placemarks {
			lineString := strings.TrimSpace(placemark.LineString)
			if len(lineString) != 0 {
				pointsStr := strings.Split(lineString, " ")
				path := make([]Point, len(pointsStr))
				for i, pointStr := range pointsStr {
					coords := strings.Split(pointStr, ",")
					if len(coords) < 2 {
						return nil, nil, fmt.Errorf("Invalid coords %s", pointsStr)
					}
					lon, err := strconv.ParseFloat(coords[0], 64)
					if err != nil {
						return nil, nil, fmt.Errorf("Can not parse y: %s", pointsStr)
					}
					lat, err := strconv.ParseFloat(coords[1], 64)
					if err != nil {
						return nil, nil, fmt.Errorf("Can not parse x: %s", pointsStr)
					}
					path[i] = Point{
						Lat:lat,
						Lon:lon,
					}
				}
				tracks = append(tracks, Track{
					Title:placemark.Name,
					Path:path,
					StartTime:JSONTime(time.Now()),
					EndTime:JSONTime(time.Now()),
				})
			}

			if len(placemark.Point) != 0 {
				point := Point3{}
				err := point.UnmarshalJSON([]byte(placemark.Point))
				if err != nil {
					logrus.Infof("Can not parse point: %s", placemark.Point)
					return nil, nil, err
				}
				points = append(points, EventPoint{
					Title: placemark.Name,
					Content: placemark.Description,
					Type: POST,
					Time: JSONTime(time.Now()),
					Point: point.toPointSwappingLatLon(),
				})
			}
		}
	}
	return tracks, points, nil
}
