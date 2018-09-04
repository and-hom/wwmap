package dao

import (
	"fmt"
	"time"
	"math"
	. "github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/model"
	"bytes"
)

type JSONTime time.Time

func (t JSONTime)MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02"))
	return []byte(stamp), nil
}

type EventPointType string;

const (
	PHOTO EventPointType = "photo"
	VIDEO EventPointType = "video"
	POST_ EventPointType = "post"
)

var EventPointAvailableTypes []EventPointType = []EventPointType{PHOTO, VIDEO, POST_}

func ParseEventPointType(s string) (EventPointType, error) {
	for _, t := range EventPointAvailableTypes {
		if s == string(t) {
			return t, nil
		}
	}
	return "", fmt.Errorf("Unsupported point type %s", s)
}

type IdTitle struct {
	Id    int64 `json:"id"`
	Title string `json:"title"`
}

type EventPoint struct {
	Id      int64 `json:"id"`
	Type    EventPointType `json:"type"`
	Point   Point `json:"point"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Time    JSONTime `json:"time"`
}

type TrackType string;

const (
	UNKNOWN TrackType = ""
	PEDESTRIAN TrackType = "pd"
	BIKE TrackType = "bk"
	WATER TrackType = "ww"
)

var TrackAvailableTypes []TrackType = []TrackType{PEDESTRIAN, BIKE, WATER, UNKNOWN}

func ParseTrackType(s string) (TrackType, error) {
	for _, t := range TrackAvailableTypes {
		if s == string(t) {
			return t, nil
		}
	}
	return "", fmt.Errorf("Unsupported track type %s", s)
}

type Track struct {
	Id        int64 `json:"id"`
	Title     string `json:"title"`
	Path      []Point `json:"path"`
	Length    float64 `json:"length"`
	Type      TrackType `json:"type"`
	StartTime JSONTime `json:"start"`
	EndTime   JSONTime `json:"end"`
}

func (this Track) Bounds() Bbox {
	if len(this.Path) == 0 {
		return Bbox{-180, -90, 180, 90}
	}
	var xMin float64 = 180
	var yMin float64 = 90
	var xMax float64 = -180
	var yMax float64 = -90

	for _, p := range this.Path {
		xMin = math.Min(xMin, p.Lat)
		yMin = math.Min(yMin, p.Lon)
		xMax = math.Max(xMax, p.Lat)
		yMax = math.Max(yMax, p.Lon)
	}

	return Bbox{
		X1:xMin,
		Y1:yMin,
		X2:xMax,
		Y2:yMax,
	}
}

type Route struct {
	Id       int64 `json:"id"`
	Title    string `json:"title"`
	Tracks   []Track `json:"tracks"`
	Points   []EventPoint `json:"points"` // points with articles
	Category model.SportCategory `json:"category"`
}

func Bounds(tracks []Track, points []EventPoint) Bbox {
	var xMin float64 = 180
	var yMin float64 = 90
	var xMax float64 = -180
	var yMax float64 = -90

	for _, tr := range tracks {
		trackBounds := tr.Bounds()
		xMin = math.Min(xMin, trackBounds.X1)
		yMin = math.Min(yMin, trackBounds.Y1)
		xMax = math.Max(xMax, trackBounds.X2)
		yMax = math.Max(yMax, trackBounds.Y2)
	}
	for _, ep := range points {
		xMin = math.Min(xMin, ep.Point.Lat)
		yMin = math.Min(yMin, ep.Point.Lon)
		xMax = math.Max(xMax, ep.Point.Lat)
		yMax = math.Max(yMax, ep.Point.Lon)
	}

	return Bbox{
		X1:xMin,
		Y1:yMin,
		X2:xMax,
		Y2:yMax,
	}
}

type ExtDataTrack struct {
	Title   string `json:"title"`
	FileIds []string `json:"fileIds"`
}

type RiverTitle struct {
	IdTitle
	OsmId    int64 `json:"osm_id"`
	RegionId int64 `json:"region_id"`
	Bounds   Bbox `json:"bounds"`
	Aliases  []string `json:"aliases"`
}

type WaterWay struct {
	Id      int64 `json:"id"`
	OsmId   int64 `json:"osm_id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Path    []Point `json:"path"`
	RiverId int64 `json:"river_id"`
	Comment string `json:"comment"`
}

type WhiteWaterPoint struct {
	IdTitle
	OsmId     int64 `json:"-"`
	RiverId   int64 `json:"river_id"`
	Type      string `json:"-"`
	Category  model.SportCategory `json:"category"`
	Point     Point `json:"point"`
	Link      string `json:"link"`
	Comment   string `json:"-"`
	ShortDesc string `json:"short_description"`
}

type WhiteWaterPointFull struct {
	WhiteWaterPoint
	LowWaterCategory       model.SportCategory `json:"lw_category"`
	LowWaterDescription    string `json:"lw_description"`
	MediumWaterCategory    model.SportCategory `json:"mw_category"`
	MediumWaterDescription string `json:"mw_description"`
	HighWaterCategory      model.SportCategory `json:"hw_category"`
	HighWaterDescription   string `json:"hw_description"`

	Orient                 string `json:"orient"`
	Approach               string `json:"approach"`
	Safety                 string `json:"safety"`

	River                  IdTitle `json:"river"`

	OrderIndex             int `json:"order_index,string"`
	AutomaticOrdering      bool `json:"automatic_ordering"`
	LastAutomaticOrdering  time.Time `json:"last_automatic_ordering"`
}

type WhiteWaterPointWithRiverTitle struct {
	WhiteWaterPoint
	RiverTitle string
	Images     []Img
}

type WhiteWaterPointWithPath struct {
	WhiteWaterPointFull
	Path []string
}

type PointRef struct {
	Id       int64 `json:"id"`
	ParentId int64 `json:"parent_id"`
	Idx      int `json:"idx"`
}

type Report struct {
	Id        int64 `json:"id"`
	ObjectId  int64 `json:"object_id,omitempty"`
	Comment   string `json:"comment"`
	CreatedAt JSONTime `json:"created_at,omitempty"`
}

type ReportWithName struct {
	Id         int64
	ObjectId   int64
	RiverTitle string
	Title      string
	Comment    string
	CreatedAt  time.Time
}

type VoyageReport struct {
	Id            int64 `json:"id,omitempty"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Source        string `json:"source"`
	RemoteId      string `json:"remote_id"`
	Url           string `json:"url"`
	DatePublished time.Time `json:"id,omitempty"`
	DateModified  time.Time `json:"id,omitempty"`
	DateOfTrip    time.Time `json:"id,omitempty"`
	Tags          []string `json:"-"`
}

type ImageType string;

const (
	IMAGE_TYPE_IMAGE ImageType = "image"
	IMAGE_TYPE_SCHEMA ImageType = "schema"
)

const IMG_SOURCE_WWMAP string = "wwmap"

type Img struct {
	Id              int64 `json:"id"`
	WwId            int64 `json:"ww_id"`
	ReportId        int64 `json:"report_id"`
	Source          string `json:"source"`
	RemoteId        string `json:"remote_id"`
	RawUrl          string `json:"-"`
	Url             string `json:"url"`
	PreviewUrl      string `json:"preview_url"`
	DatePublished   time.Time `json:"date_published"`
	LabelsForSearch []string `json:"-"`
	Enabled         bool `json:"enabled"`
	Type            ImageType `json:"type"`
	MainImage       bool        `json:"main_image"`
}

func GetImgType(_type string) ImageType {
	if t, f := checkType(_type, IMAGE_TYPE_IMAGE); f {
		return t
	}
	if t, f := checkType(_type, IMAGE_TYPE_SCHEMA); f {
		return t
	}
	return IMAGE_TYPE_IMAGE
}

func checkType(val string, _type ImageType) (ImageType, bool) {
	if val == string(_type) {
		return _type, true
	}
	return IMAGE_TYPE_IMAGE, false
}

func (this Img) IdStr() string {
	return fmt.Sprintf("%d", this.Id)
}

type WWPassport struct {
	Source        string
	RemoteId      string
	WwId          int64
	Url           string
	DatePublished time.Time
	DateModified  time.Time
	River         string
	Title         string
}

type Role string

const (
	ADMIN Role = "ADMIN"
	EDITOR Role = "EDITOR"
	USER Role = "USER"
	ANONYMOUS Role = "ANON"
)

func Join(separator string, roles ...Role) string {
	if len(roles) == 1 {
		return string(roles[0])
	}
	if len(roles) == 2 {
		return string(roles[0]) + separator + string(roles[1])
	}
	var buffer bytes.Buffer
	for i := 0; i < len(roles); i++ {
		if i > 0 {
			buffer.WriteString(separator)
		}
		buffer.WriteString(string(roles[i]))
	}
	return buffer.String()
}

type UserInfo struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Login     string `json:"login"`
}

type User struct {
	Id       int64 `json:"id"`
	YandexId int64 `json:"yandex_id"`
	Role     Role `json:"role"`
	Info     UserInfo `json:"info"`
}

type Country struct {
	Id    int64 `json:"id,omitempty"`
	Title string `json:"title"`
}

type Region struct {
	Id        int64 `json:"id,omitempty"`
	CountryId int64 `json:"country_id,omitempty"`
	Title     string `json:"title"`
}

type RegionWithCountry struct {
	Id      int64 `json:"id,omitempty"`
	Country Country `json:"country,omitempty"`
	Title   string `json:"title"`
}
