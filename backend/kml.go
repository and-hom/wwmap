package main

import (
	"io"
	"encoding/xml"
	"io/ioutil"
	"strings"
	"fmt"
	"strconv"
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

func (this KmlParser) getTracks() ([]Track, error) {
	tracks := make([]Track, 0)

	for _, folder := range this.kml_data.Document.Folders {
		for _, placemark := range folder.Placemarks {
			lineString := strings.TrimSpace(placemark.LineString)
			if len(lineString) != 0 {
				pointsStr := strings.Split(lineString, " ")
				path := make([]Point, len(pointsStr))
				for i, pointStr := range pointsStr {
					coords := strings.Split(pointStr, ",")
					if len(coords) < 2 {
						return nil, fmt.Errorf("Invalid coords %s", pointsStr)
					}
					x, err := strconv.ParseFloat(coords[1], 64)
					if err != nil {
						return nil, fmt.Errorf("Can not parse x: %s", pointsStr)
					}
					y, err := strconv.ParseFloat(coords[0], 64)
					if err != nil {
						return nil, fmt.Errorf("Can not parse y: %s", pointsStr)
					}
					path[i] = Point{
						lat:x,
						lon:y,
					}
				}
				tracks = append(tracks, Track{
					Title:placemark.Name,
					Path:path,
				})
			}
		}
	}
	return tracks, nil
}
