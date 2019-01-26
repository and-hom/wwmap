package dao

import (
	"bytes"
	"fmt"
	. "github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/model"
	"math"
	"strconv"
	"time"
)

type JSONDate time.Time

func (t JSONDate) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", t.String())
	return []byte(stamp), nil
}

func (t JSONDate) String() string {
	return time.Time(t).Format("2006-01-02")
}

type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", t.String())
	return []byte(stamp), nil
}

func (t JSONTime) String() string {
	return time.Time(t).Format("2006-01-02 00:00")
}

type JSONUnixTime time.Time

func (t JSONUnixTime) MarshalJSON() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t JSONUnixTime) String() string {
	return fmt.Sprintf("%d", time.Time(t).Unix())
}

func (this *JSONUnixTime) UnmarshalJSON(data []byte) error {
	t, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*this = JSONUnixTime(time.Unix(t, 0))
	return nil
}

type IdTitle struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
}

type SpotCounters struct {
	Ordered int `json:"ordered"`
	Total   int `json:"total"`
}

type RiverTitle struct {
	IdTitle
	OsmId   int64                  `json:"osm_id"`
	Region  Region                 `json:"region"`
	Bounds  Bbox                   `json:"bounds"`
	Aliases []string               `json:"aliases"`
	Props   map[string]interface{} `json:"props"`
	Visible bool                   `json:"visible"`
}

func (this RiverTitle) GetId() int64 {
	return this.Id
}
func (this RiverTitle) GetProperties() map[string]interface{} {
	return this.Props
}

type River struct {
	RiverTitle
	Description  string       `json:"description"`
	SpotCounters SpotCounters `json:"spot_counters"`
}

type Spot struct {
	IdTitle
	Point       Point
	Category    model.SportCategory
	Description string
	Images      []Img
	Link        string
	Props       map[string]interface{}
}

type RiverWithSpots struct {
	IdTitle
	Spots     []Spot
	RegionId  int64
	CountryId int64
}

type RiverWithSpotsExt struct {
	RiverWithSpots
	Description string
	Props       map[string]interface{}
	Region      Region
}

func (this RiverWithSpotsExt) GetId() int64 {
	return this.Id
}
func (this RiverWithSpotsExt) GetProperties() map[string]interface{} {
	return this.Props
}

type WaterWay struct {
	Id      int64   `json:"id"`
	OsmId   int64   `json:"osm_id"`
	Title   string  `json:"title"`
	Type    string  `json:"type"`
	Path    []Point `json:"path"`
	RiverId int64   `json:"river_id"`
	Comment string  `json:"comment"`
}

const EXPORT_PROP_PREFIX = "export_"
const PAGE_LINK_PROP_PREFIX = "page_link_"

type WhiteWaterPoint struct {
	IdTitle
	OsmId     int64               `json:"-"`
	RiverId   int64               `json:"river_id"`
	Type      string              `json:"-"`
	Category  model.SportCategory `json:"category"`
	Point     Point               `json:"point"`
	Link      string              `json:"link"`
	Comment   string              `json:"-"`
	ShortDesc string              `json:"short_description"`
}

type WhiteWaterPointFull struct {
	WhiteWaterPoint
	LowWaterCategory       model.SportCategory `json:"lw_category"`
	LowWaterDescription    string              `json:"lw_description"`
	MediumWaterCategory    model.SportCategory `json:"mw_category"`
	MediumWaterDescription string              `json:"mw_description"`
	HighWaterCategory      model.SportCategory `json:"hw_category"`
	HighWaterDescription   string              `json:"hw_description"`

	Orient   string `json:"orient"`
	Approach string `json:"approach"`
	Safety   string `json:"safety"`

	River IdTitle `json:"river"`

	OrderIndex            int       `json:"order_index,string"`
	AutomaticOrdering     bool      `json:"automatic_ordering"`
	LastAutomaticOrdering time.Time `json:"last_automatic_ordering"`

	Aliases []string               `json:"aliases"`
	Props   map[string]interface{} `json:"props"`
}

type WhiteWaterPointWithRiverTitle struct {
	WhiteWaterPoint
	RiverTitle string `json:"river_title"`
	Images     []Img  `json:"images"`
}

type PointRef struct {
	Id       int64 `json:"id"`
	ParentId int64 `json:"parent_id"`
	Idx      int   `json:"idx"`
}

type NotificationProvider string

const NOTIFICATION_PROVIDER_LOG NotificationProvider = "log"
const NOTIFICATION_PROVIDER_EMAIL NotificationProvider = "email"
const NOTIFICATION_PROVIDER_VK NotificationProvider = "vk"

type Notification struct {
	IdTitle

	Object    IdTitle
	Comment   string
	CreatedAt JSONDate

	Recipient  NotificationRecipient
	Classifier string
	SendBefore time.Time
}

type NotificationRecipient struct {
	Provider  NotificationProvider
	Recipient string
}

func (this NotificationRecipient) String() string {
	return fmt.Sprintf("%s/%s", this.Provider, this.Recipient)
}

type NotificationRecipientWithClassifier struct {
	NotificationRecipient
	Classifier string
}

func (this NotificationRecipientWithClassifier) String() string {
	return fmt.Sprintf("%v - %s", this.NotificationRecipient, this.Classifier)
}

type VoyageReport struct {
	Id            int64     `json:"id,omitempty"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	Source        string    `json:"source"`
	RemoteId      string    `json:"remote_id"`
	Url           string    `json:"url"`
	DatePublished time.Time `json:"id,omitempty"`
	DateModified  time.Time `json:"id,omitempty"`
	DateOfTrip    time.Time `json:"id,omitempty"`
	Tags          []string  `json:"-"`
}

type ImageType string;

const (
	IMAGE_TYPE_IMAGE  ImageType = "image"
	IMAGE_TYPE_SCHEMA ImageType = "schema"
)

const IMG_SOURCE_WWMAP string = "wwmap"

type Img struct {
	Id              int64     `json:"id"`
	WwId            int64     `json:"ww_id"`
	ReportId        int64     `json:"report_id"`
	Source          string    `json:"source"`
	RemoteId        string    `json:"remote_id"`
	RawUrl          string    `json:"-"`
	Url             string    `json:"url"`
	PreviewUrl      string    `json:"preview_url"`
	DatePublished   time.Time `json:"date_published"`
	LabelsForSearch []string  `json:"-"`
	Enabled         bool      `json:"enabled"`
	Type            ImageType `json:"type"`
	MainImage       bool      `json:"main_image"`
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
	ADMIN     Role = "ADMIN"
	EDITOR    Role = "EDITOR"
	USER      Role = "USER"
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

type AuthProvider string

const YANDEX AuthProvider = "yandex"
const GOOGLE AuthProvider = "google"
const VK AuthProvider = "vk"

func (this AuthProvider) HumanName() string {
	switch this {
	case YANDEX:
		return "Яндекс"
	case GOOGLE:
		return "Google"
	case VK:
		return "ВКонтакте"
	default:
		return string(this)
	}
}

type User struct {
	Id           int64        `json:"id"`
	ExtId        string       `json:"ext_id"`
	AuthProvider AuthProvider `json:"auth_provider"`
	Role         Role         `json:"role"`
	Info         UserInfo     `json:"info"`
	SessionId    string       `json:"session_id"`
}

type Country struct {
	Id    int64  `json:"id,omitempty"`
	Title string `json:"title"`
	Code  string `json:"code"`
}

type Region struct {
	Id        int64  `json:"id,omitempty"`
	CountryId int64  `json:"country_id,omitempty"`
	Title     string `json:"title"`
	Fake      bool   `json:"fake,omitempty"`
}

type RegionWithCountry struct {
	Id      int64   `json:"id,omitempty"`
	Country Country `json:"country,omitempty"`
	Title   string  `json:"title"`
	Fake    bool    `json:"fake,omitempty"`
}

type ChangesLogEntry struct {
	Id           int64               `json:"id,omitempty"`
	ObjectType   string              `json:"object_type"`
	ObjectId     int64               `json:"object_id"`
	AuthProvider AuthProvider        `json:"auth_provider"`
	ExtId        string              `json:"ext_id"`
	Login        string              `json:"login"`
	Type         ChangesLogEntryType `json:"type"`
	Description  string              `json:"description"`
	Time         JSONTime            `json:"time"`
}

type Daytime string

const (
	NIGHT   Daytime = "N"
	MORNING Daytime = "M"
	DAY     Daytime = "D"
	EVENING Daytime = "E"
)

type MeteoPoint struct {
	IdTitle
	Point       Point `json:"point"`
	CollectData bool  `json:"-"`
}

type Meteo struct {
	Id      int64 `json:"id,omitempty"`
	PointId int64 `json:"point_id"`

	Date    JSONDate `json:"date"`
	Daytime Daytime  `json:"daytime"`

	Temp int `json:"temp"`
	Rain int `json:"rain"`
}

const NAN_LEVEL = math.MinInt32

type Level struct {
	Id        int64    `json:"id,omitempty"`
	SensorId  string   `json:"sensor_id"`
	Date      JSONDate `json:"date"`
	HourOfDay int16    `json:"hour_of_day"`
	Level     int      `json:"level"`
}

type ChangesLogEntryType string

const (
	ENTRY_TYPE_CREATE ChangesLogEntryType = "CREATE"
	ENTRY_TYPE_MODIFY ChangesLogEntryType = "MODIFY"
	ENTRY_TYPE_DELETE ChangesLogEntryType = "DELETE"
)

const CATEGORY_DEFINITING_POINTS_COUNT int = 3
const MAX_CATEGORY = 6

func CalculateClusterCategory(points []Spot) int {
	cntByCat := make(map[int]int)
	categorizedPointsCount := 0
	for i := 0; i < len(points); i++ {
		currentCat := points[i].Category.Category
		cntByCat[currentCat] += 1
		if currentCat > 0 {
			categorizedPointsCount += 1
		}
	}

	wwCnt := 0
	riverCategory := 0
	definitingPointsCount := min(CATEGORY_DEFINITING_POINTS_COUNT, categorizedPointsCount)
	for i := MAX_CATEGORY; i > 0 && wwCnt < definitingPointsCount; i-- {
		wwCnt += cntByCat[i]
		riverCategory = i
	}
	return riverCategory
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

type PgPoint struct {
	Coordinates Point `json:"coordinates"`
}

func (this PgPoint) GetPoint() Point {
	// flip coordinates for postGIS
	return this.Coordinates.Flip()
}

type PgPolygon struct {
	Coordinates [][]Point `json:"coordinates"`
}
