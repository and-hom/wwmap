package handler

import (
	"fmt"
	"github.com/and-hom/wwmap/lib/dao"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"net/http"
	"time"
)

const PLOT_DAYS = 10

type DashboardHandler struct {
	App
	LevelDao dao.LevelDao
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
	this.JsonAnswer(w, refs)
}

func (this *DashboardHandler) Levels(w http.ResponseWriter, req *http.Request) {
	rivers, err := this.RiverDao.ListAll()
	if err != nil {
		OnError500(w, err, "Can not get all rivers")
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

	today := time.Now()
	_10daysLevels := today.Add(time.Hour * 24 * (-PLOT_DAYS))
	levelData, err := this.LevelDao.List(dao.JSONDate(_10daysLevels))
	if err != nil {
		OnError500(w, err, "Can't list sensor data")
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

		labels := make([]string, PLOT_DAYS)
		line := JChartDataSet{
			BackgroundColor: []string{"blue"},
			BorderColor:     []string{"blue"},
			Fill:            false,
		}
		for i := int64(0); i < PLOT_DAYS; i++ {
			hoursOffset := time.Duration(int64(time.Hour) * 24 * (1 + i - PLOT_DAYS))
			date := dao.JSONDate(today.Add(hoursOffset))
			labels[i] = date.String()
			var levelValue *int
			for j := 0; j < len(data); j++ {
				if data[j].Date.String() == date.String() {
					levelValue = &data[j].Level
					break
				}
			}
			line.Data = append(line.Data, levelValue)
		}

		sensorData := JChartData{
			Labels:   labels,
			DataSets: []JChartDataSet{line},
		}

		result[sensorId] = SensorData{
			SensorId:  sensorId,
			Rivers:    rivers,
			ChartData: sensorData,
		}
	}

	this.JsonAnswer(w, result)
}

type SensorData struct {
	SensorId  string               `json:"sensor_id"`
	Rivers    []RiverWithRegionDto `json:"rivers"`
	ChartData JChartData           `json:"chart_data"`
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
	Fill            bool     `json:"fill"`
}

type RiverWithRegionDto struct {
	dao.IdTitle
	Region dao.Region `json:"region"`
}
