package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/cron/vodinfo-eye/graduation"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/pkg/errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"
	"time"
)

const URL_TEMPLATE = "http://gis.vodinfo.ru/informer/draw/v2_%s_400_300_10_ffffff_110_8_7_H_none.png"
const (
	X_LEVEL_VALUE_AREA     = 44
	Y_LEVEL_VALUE_AREA_MIN = 40
	Y_LEVEL_VALUE_AREA_MAX = 240
)

func main() {
	log.Infof("Starting wwmap vodinfo import")
	configuration := config.Load("")
	configuration.LogLevel = "info"
	configuration.ChangeLogLevel()
	storage := dao.NewPostgresStorage(configuration.Db)
	graduator, err := graduation.NewPercentileGladiator(0.1, 0.1)
	if err != nil {
		log.Fatal("Can't initialize graduator: ", err)
	}

	patternMatcher, err := NewPatternMatcher()
	if err != nil {
		log.Fatal("Failed to load level value number patterns: ", err)
	}

	VodinfoLevelRetriever{
		configuration:  configuration,
		riverDao:       dao.NewRiverPostgresDao(storage),
		levelDao:       dao.NewLevelPostgresDao(storage),
		levelSensorDao: dao.NewLevelSensorPostgresDao(storage),
		graduator:      graduator,
		client: http.Client{
			Timeout: 4 * time.Second,
		},
		patternMatcher: patternMatcher,
		backScanDays:   9,
	}.doRetrieveLevel()
}

type VodinfoLevelRetriever struct {
	configuration  config.Configuration
	riverDao       dao.RiverDao
	levelDao       dao.LevelDao
	levelSensorDao dao.LevelSensorDao
	graduator      graduation.LevelGraduator
	client         http.Client
	patternMatcher PatternMatcher
	backScanDays   int
}

func (this VodinfoLevelRetriever) listSensors() []string {
	log.Info("List used level sensors")
	rivers, err := this.riverDao.ListAll()
	if err != nil {
		log.Fatal("Failed to list rivers: ", err)
	}

	sensorIdMap := make(map[string]bool)
	for _, river := range rivers {
		sensorIdArr, exists := river.Props["vodinfo_sensors"]
		if !exists {
			continue
		}
		fmt.Println(sensorIdArr)
		for _, id := range sensorIdArr.([]interface{}) {
			sensorId := strconv.Itoa(int(id.(float64)))
			sensorIdMap[sensorId] = true
		}
	}

	storedSensors, err := this.levelSensorDao.List()
	if err != nil {
		log.Fatal("Can't list sensors in db: ", err)
	}
	for _, s := range storedSensors {
		sensorIdMap[s.Id] = true
	}

	sensorIds := make([]string, 0, len(sensorIdMap))
	for k, _ := range sensorIdMap {
		sensorIds = append(sensorIds, k)
	}
	fmt.Println(storedSensors)
	log.Info("SensorIds: ", sensorIds)

	return sensorIds
}

func (this VodinfoLevelRetriever) doRetrieveLevel() {
	sensorIds := this.listSensors()

	t1 := util.ToDateInDefaultZone(time.Now().Add(-24 * time.Duration(this.backScanDays) * time.Hour))
	t2 := util.ToDateInDefaultZone(time.Now())
	previousDaysLevelsBySensor, err := this.levelDao.ListBySensorAndDate(t1, t2)
	if err != nil {
		log.Warn("Can't find yesterday levels", err)
		previousDaysLevelsBySensor = make(map[string]map[string]dao.Level)
	}

	for _, sensorId := range sensorIds {
		log.Infof("Try to sensor %s row if missing", sensorId)
		err = this.levelSensorDao.CreateIfMissing(sensorId)
		if err != nil {
			log.Warn("Can't check level sensor and create if missing ", err)
		}

		log.Infof("Fetch and calibrate level data for sensor %s", sensorId)
		todayLevel := dao.NAN_LEVEL
		calibrated, err := LoadImage(sensorId, &this.client, &this.patternMatcher)
		if err != nil {
			log.Error(err)
			continue
		}

		if !calibrated.Ok {
			log.Errorf("Image calibration for sensor %s failed", sensorId)
			continue
		}

		log.Infof("Detect and save today level value for sensor %s", sensorId)
		todayLevel = calibrated.GetLevelValue(DetectLine)
		now := time.Now().In(util.GetDefaultLocation())
		err = this.levelDao.Insert(dao.Level{
			SensorId:  sensorId,
			Date:      dao.JSONDate(now.Truncate(24 * time.Hour)),
			HourOfDay: int16(now.Hour()),
			Level:     todayLevel,
		})
		if err != nil {
			log.Errorf("Can't insert level value for %s: %v", sensorId, err)
			continue
		}

		log.Infof("Detect missing previous days data for sensor %s", sensorId)
		dailyLevels, found := previousDaysLevelsBySensor[sensorId]
		if !found {
			dailyLevels = make(map[string]dao.Level)
		}
		log.Warn("No prev days level data for sensor ", sensorId)
		for dayOffset := 1; dayOffset <= this.backScanDays; dayOffset++ {

			day := util.ToDateInDefaultZone(time.Now().Add(-24 * time.Duration(dayOffset) * time.Hour))
			dayStr := util.FormatDate(day)
			_, found := dailyLevels[dayStr]
			if found {
				log.Infof("Level for %s %s already exists", sensorId, dayStr)
				continue
			}

			log.Infof("Detect missing level for sensor %s for today-%ddays (%s)", sensorId, dayOffset, day.String())
			if levelValue := calibrated.GetLevelValue(func(img image.Image) int {
				return DetectPrevDaysLine(img, dayOffset)
			}); levelValue != dao.NAN_LEVEL {
				err = this.levelDao.Insert(dao.Level{
					SensorId:  sensorId,
					Date:      dao.JSONDate(day),
					HourOfDay: 25,
					Level:     levelValue,
				})
				if err != nil {
					log.Errorf("Can't insert level value for %s: %v", sensorId, err)
					continue
				}
			}
		}

		log.Infof("Re-graduate sensor %s", sensorId)
		graduation.ReCalculateSensorMinMax(this.graduator, this.levelSensorDao, this.levelDao, sensorId)
	}
}

func LoadImage(sensorId string, client *http.Client, patternMatcher *PatternMatcher) (CalibratedImage, error) {
	img, err := DownloadImage(sensorId, client)
	if err != nil {
		return CalibratedImage{Ok: false}, err
	}
	imgCData, err := Calibrate(img, patternMatcher)
	if err != nil {
		return CalibratedImage{Ok: false}, err
	}
	return CalibratedImage{CalibrationData: imgCData, Img: img, Ok: true}, nil
}

type CalibratedImage struct {
	Img             image.Image
	CalibrationData ImageCalibrationData
	Ok              bool
}

func DownloadImage(sensorId string, client *http.Client) (image.Image, error) {
	log.Infof("Read informer for %s", sensorId)
	url := fmt.Sprintf(URL_TEMPLATE, sensorId)
	log.Infof("Download image %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Can't create request for %s: %v", url, err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/601.7.7 (KHTML, like Gecko) Version/9.1.2 Safari/601.7.7")
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Can't perform request for %s: %v", url, err)
		return nil, err
	}
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Errorf("Can't decode image for %s: %v", url, err)
		return nil, err
	}
	return img, nil
}

func Calibrate(img image.Image, matcher *PatternMatcher) (ImageCalibrationData, error) {
	yAxisLabelsCoords := (*matcher).Match(img, X_LEVEL_VALUE_AREA)
	log.Debug("Y axis labels coords: ", yAxisLabelsCoords)
	if len(yAxisLabelsCoords) == 0 {
		return ImageCalibrationData{}, errors.New("No labels detected - can't process")
	}
	if len(yAxisLabelsCoords) == 1 {
		return ImageCalibrationData{}, errors.New("Single label detected - can't determine scale")
	}
	yAxisMarksXCoords, err := DetectYAxisLabels(img, yAxisLabelsCoords)
	if err != nil {
		return ImageCalibrationData{}, fmt.Errorf("Can't detect y axis: %v", err)
	}
	log.Info("Y axis marks coords: ", yAxisMarksXCoords)
	imgCData := minAndMaxLevelAndYVal(yAxisMarksXCoords)
	log.Info(imgCData.String())
	return imgCData, nil
}

func (this CalibratedImage) GetLevelValue(detectLevelY func(image.Image) int) int {
	y := detectLevelY(this.Img)
	if y < 0 {
		log.Errorf("Can't detect plot line")
		return dao.NAN_LEVEL
	}
	level := this.CalibrationData.YToLevel(float64(y))
	log.Debug("Level is ", level)
	return int(level)
}

func minAndMaxLevelAndYVal(yAxisMarksXCoords map[int]int) ImageCalibrationData {
	result := InitialImageCalibrationData()
	for l, y := range yAxisMarksXCoords {
		result.Add(float64(l), float64(y))
	}
	return result
}
