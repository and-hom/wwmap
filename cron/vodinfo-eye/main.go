package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"time"
)

const URL_TEMPLATE = "http://gis.vodinfo.ru/informer/draw/v2_%d_400_300_10_ffffff_110_8_7_H_none.png"
const (
	X_LEVEL_VALUE_AREA     = 44
	Y_LEVEL_VALUE_AREA_MIN = 40
	Y_LEVEL_VALUE_AREA_MAX = 240
)

func main() {
	log.Infof("Starting wwmap vodinfo import")
	configuration := config.Load("")
	configuration.ChangeLogLevel()
	storage := dao.NewPostgresStorage(configuration.Db)

	riverDao := dao.NewRiverPostgresDao(storage)
	levelDao := dao.NewLevelPostgresDao(storage)

	client := http.Client{
		Timeout: 4 * time.Second,
	}

	rivers, err := riverDao.ListAll()
	if err != nil {
		log.Fatal("Failed to list rivers: ", err)
	}

	patternMatcher, err := NewPatternMatcher()
	if err != nil {
		log.Fatal("Failed to load level value number patterns: ", err)
	}

	sensorIds := make(map[int]bool)
	for _, river := range rivers {
		sensorIdF, exists := river.Props["vodinfo_sensor"]
		if !exists {
			continue
		}
		sensorId := int(sensorIdF.(float64))
		sensorIds[sensorId] = true
	}

	for sensorId, _ := range sensorIds {
		lToday := dao.NAN_LEVEL
		img, err := DownloadImage(sensorId, client)
		if err == nil {
			lToday = GetLevelValue(img, patternMatcher)
		}

		err = levelDao.Insert(dao.Level{
			SensorId: fmt.Sprintf("%d", sensorId),
			Date:     dao.JSONDate(time.Now()),
			Level:    lToday,
		})
		if err != nil {
			log.Errorf("Can't insert level value for %d: %v", sensorId, err)
			continue
		}
	}
}

func DownloadImage(sensorId int, client http.Client) (image.Image, error) {
	log.Infof("Read informer for %d", sensorId)
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

func GetLevelValue(img image.Image, matcher PatternMatcher) int {
	yAxisLabelsCoords := matcher.Match(img, X_LEVEL_VALUE_AREA)
	log.Debug("Y axis labels coords: ", yAxisLabelsCoords)
	if len(yAxisLabelsCoords) == 0 {
		log.Errorf("No labels detected - can't process")
		return dao.NAN_LEVEL
	}
	if len(yAxisLabelsCoords) == 1 {
		log.Errorf("Single label detected - can't determine scale")
		return dao.NAN_LEVEL
	}
	yAxisMarksXCoords, err := DetectYAxisLabels(img, yAxisLabelsCoords)
	if err != nil {
		log.Errorf("Can't detect y axis: %v", err)
		return dao.NAN_LEVEL
	}
	log.Info("Y axis marks coords: ", yAxisMarksXCoords)
	lMin, yMin, lMax, yMax := minAndMaxLevelAndYVal(yAxisMarksXCoords)
	log.Infof("LMin=%f YMin=%f LMax=%f YMax=%f", lMin, yMin, lMax, yMax)

	yToday := DetectLine(img)
	if yToday < 0 {
		log.Errorf("Can't detect plot line")
		return dao.NAN_LEVEL
	}
	lToday := (yMin-float64(yToday))/(yMin-yMax)*(lMax-lMin) + lMin
	log.Debug("Level is ", lToday)
	return int(lToday)
}

func minAndMaxLevelAndYVal(yAxisMarksXCoords map[int]int) (float64, float64, float64, float64) {
	lMax := -100000.0
	yMax := -100000.0
	lMin := 100000.0
	yMin := 100000.0
	for l, y := range yAxisMarksXCoords {
		lF := float64(l)
		if lF > lMax {
			lMax = lF
			yMax = float64(y)
		}
		if lF < lMin {
			lMin = lF
			yMin = float64(y)
		}
	}
	return lMin, yMin, lMax, yMax
}
