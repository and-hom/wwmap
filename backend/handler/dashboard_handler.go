package handler

import (
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/and-hom/wwmap/lib/util"
	"net/http"
	"time"
)

const DEFAULT_PLOT_DAYS = 10

const GRAD_COLOR_C5 = "#0055bb80"
const GRAD_COLOR_C4 = "#0055bb65"
const GRAD_COLOR_C3 = "#0055bb50"
const GRAD_COLOR_C2 = "#0055bb35"
const GRAD_COLOR_C1 = "#0055bb20"

type DashboardHandler struct {
	App
	LevelDao       dao.LevelDao
	LevelSensorDao dao.LevelSensorDao
}

func (this *DashboardHandler) Init() {
	this.Register("/dashboard/ref-sites", HandlerFunctions{Get: this.RefSites})
	this.Register("/dashboard/levels", HandlerFunctions{Get: this.Levels})
}

func (this *DashboardHandler) RefSites(w http.ResponseWriter, req *http.Request) {
	SetJsonResponseHeaders(w)

	refs, err := this.RefererStorage.List()
	if err != nil {
		OnError500(w, err, "Can not list referers")
		return
	}
	JsonAnswer(w, refs)
}

func (this *DashboardHandler) Levels(w http.ResponseWriter, req *http.Request) {
	rivers, err := this.RiverDao.ListAll()
	if err != nil {
		OnError500(w, err, "Can not get all rivers")
		return
	}

	riversBySensor := make(map[string][]dao.RiverTitle)
	for _, river := range rivers {
		sensorIds, exists := river.Props["vodinfo_sensors"]
		if !exists {
			continue
		}
		for _, sensorIdF := range sensorIds.([]interface{}) {
			sensorId := fmt.Sprintf("%d", int(sensorIdF.(float64)))
			riversBySensor[sensorId] = append(riversBySensor[sensorId], river)
		}
	}

	fromDate, err := parseDate("from", req, -DEFAULT_PLOT_DAYS)
	if err != nil {
		OnError(w, err, "Can't parse 'from' date", http.StatusBadRequest)
		return
	}

	toDate, err := parseDate("to", req, 0)
	if err != nil {
		OnError(w, err, "Can't parse 'to' date", http.StatusBadRequest)
		return
	}

	if fromDate.After(toDate) {
		OnError(w, err, fmt.Sprintf("fromDate %s is after toDate %s", fromDate, toDate), http.StatusBadRequest)
		return
	}

	days := int64(toDate.Sub(fromDate).Hours()/24 + 1)

	levelData, err := this.LevelDao.ListBySensorAndDate(fromDate, toDate)
	if err != nil {
		OnError500(w, err, "Can't list sensor data")
		return
	}

	levelSensors, err := this.LevelSensorDao.List()
	if err != nil {
		OnError500(w, err, "Can't list level sensors")
		return
	}

	result := make(map[string]SensorData)
	for sensorId, r := range riversBySensor {
		data := levelData[sensorId]
		rivers := make([]RiverWithRegionDto, len(r))
		for i := 0; i < len(r); i++ {
			rivers[i] = RiverWithRegionDto{
				IdTitle: r[i].IdTitle,
				Region:  r[i].Region,
			}
		}

		labels := make([]string, days)
		line := JChartDataSet{
			BackgroundColor: []string{"blue"},
			BorderColor:     []string{"blue"},
			Fill:            "false",
			BorderWidth:     3,
		}
		for i := int64(0); i < days; i++ {
			hoursOffset := time.Duration(int64(time.Hour) * 24 * (1 + i - days))
			date := util.ToDateInDefaultZone(toDate.Add(hoursOffset))
			dateStr := util.FormatDate(date)

			var levelValue *int = nil
			l, found := data[dateStr]
			if found {
				levelValue = &l.Level
			}

			labels[i] = dateStr
			line.Data = append(line.Data, levelValue)
		}

		sensorMetrics := SensorMetrics{}
		for i := 0; i < len(levelSensors); i++ {
			if levelSensors[i].Id == sensorId {
				sensorMetrics.L0 = levelSensors[i].L[0]
				sensorMetrics.L1 = levelSensors[i].L[1]
				sensorMetrics.L2 = levelSensors[i].L[2]
				sensorMetrics.L3 = levelSensors[i].L[3]
				break
			}
		}

		sensorData := JChartData{
			Labels: labels,
			DataSets: []JChartDataSet{
				line,
				graduationLine(days, 1, sensorMetrics.L0, GRAD_COLOR_C1),
				graduationLine(days, 2, sensorMetrics.L1, GRAD_COLOR_C2),
				graduationLine(days, 3, sensorMetrics.L2, GRAD_COLOR_C3),
				graduationLine(days, 4, sensorMetrics.L3, GRAD_COLOR_C4),
				graduationLine(days, 4, 1000, GRAD_COLOR_C5),
			},
		}

		result[sensorId] = SensorData{
			SensorId:      sensorId,
			Rivers:        rivers,
			ChartData:     sensorData,
			SensorMetrics: sensorMetrics,
		}
	}

	JsonAnswer(w, result)
}

func graduationLine(days int64, grade int, y int, color string) JChartDataSet {
	fill := "-1"
	if grade == 1 {
		fill = "origin"
	}
	line := JChartDataSet{
		BackgroundColor: []string{color},
		BorderColor:     []string{color},
		Fill:            fill,
		PointRadius:     0,
		BorderWidth:     0,
		Label:           fmt.Sprintf("l %d", grade),
	}

	for i := int64(0); i < days; i++ {
		line.Data = append(line.Data, &y)
	}

	return line
}

func parseDate(paramName string, req *http.Request, defaultOffsetDays int) (time.Time, error) {
	toDateParam := req.FormValue(paramName)
	if toDateParam != "" {
		return time.Parse(util.DATE_FORMAT, toDateParam)
	} else {
		return time.Now().Add(time.Duration(defaultOffsetDays) * 24 * time.Hour), nil
	}
}

type SensorData struct {
	SensorId      string               `json:"sensor_id"`
	Rivers        []RiverWithRegionDto `json:"rivers"`
	ChartData     JChartData           `json:"chart_data"`
	SensorMetrics SensorMetrics        `json:"sensor_metrics"`
}

type SensorMetrics struct {
	L0 int `json:"l0"`
	L1 int `json:"l1"`
	L2 int `json:"l2"`
	L3 int `json:"l3"`
}

type JChartData struct {
	Labels   []string        `json:"labels"`
	DataSets []JChartDataSet `json:"datasets"`
}

type JChartDataSet struct {
	Label           string   `json:"label"`
	BackgroundColor []string `json:"backgroundColor"`
	BorderColor     []string `json:"borderColor"`
	Data            []*int   `json:"data"`
	Fill            string   `json:"fill"`
	PointRadius     int      `json:"pointRadius"`
	BorderWidth     int      `json:"borderWidth"`
}

type RiverWithRegionDto struct {
	dao.IdTitle
	Region dao.Region `json:"region"`
}
