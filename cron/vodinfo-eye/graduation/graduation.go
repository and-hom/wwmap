package graduation

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"time"
)

func ReCalculateSensorMinMax(graduator LevelGraduator,
	levelSensorDao dao.LevelSensorDao,
	levelDao dao.LevelDao,
	sensorId string) {

	vals, err := levelDao.ListForSensor(sensorId)
	if err != nil {
		log.Errorf("Can't select level values for sensor %s: %v", sensorId, err)
		return
	}

	graduation, err := graduator.Graduate(vals)
	if err != nil {
		log.Errorf("Can't graduate level values for sensor %s: %v", sensorId, err)
		return
	}

	err = levelSensorDao.SetGraduation(sensorId, graduation)
	if err != nil {
		log.Errorf("Can't graduate level values for sensor %s: %v", sensorId, err)
		return
	}
}

const (
	NO_LEVEL_FOR_DATE int8 = -1
	NO_SENSOR_DATA    int8 = -2
)
const day = 24 * time.Hour

func NewGraduatedLevelValuesBySensor() GraduatedLevelValuesBySensor {
	return GraduatedLevelValuesBySensor{
		Value: make(map[string]int8),
	}
}

type GraduatedLevelValuesBySensor struct {
	Value map[string]int8
}

func (this *GraduatedLevelValuesBySensor) Set(sensorId string, level int) {
	this.SetU(sensorId, int8(level))
}

func (this *GraduatedLevelValuesBySensor) SetU(sensorId string, level int8) {
	this.Value[sensorId] = level
}

func GetLevelBySensors(levelSensorDao dao.LevelSensorDao,
	levelDao dao.LevelDao, sensorIds []string,
	date time.Time, daysAround int, manualLevel int8) map[string]int8 {
	_daysAround := time.Duration(daysAround)

	levelGraduatedResult := NewGraduatedLevelValuesBySensor()
	if manualLevel >= 0 {
		levelGraduatedResult.SetU(dao.IMG_WATER_LEVEL_MANUAL, manualLevel)
	}
	if date.Year() <= 1 {
		return levelGraduatedResult.Value
	}
	for _, sensorId := range sensorIds {
		s, err := levelSensorDao.Find(sensorId)
		if err != nil {
			levelGraduatedResult.SetU(sensorId, NO_SENSOR_DATA)
			log.Errorf("Can't select sensor %d: %v", s, err)
			continue
		}
		levels, err := levelDao.GetDailyLevelBetweenDates(sensorId,
			date.Add(-_daysAround*day),
			date.Add(_daysAround*day))
		if err != nil {
			levelGraduatedResult.SetU(sensorId, NO_SENSOR_DATA)
			log.Errorf("Can't select level for sensor %d and date %s: %v", s, date.String(), err)
			continue
		}
		if len(levels) == 0 {
			levelGraduatedResult.SetU(sensorId, NO_LEVEL_FOR_DATE)
			continue
		}

		levelValue := getLevelValue(levels, date)

		log.Debugf("Level for sensor %s on %s is %d", sensorId, date.String(), levelValue)
		for graduatedLevel := 0; graduatedLevel < dao.LEVEL_GRADUATION; graduatedLevel++ {
			if levelValue < s.L[graduatedLevel] {
				levelGraduatedResult.Set(sensorId, graduatedLevel+1)
				break
			}
		}
	}

	return levelGraduatedResult.Value
}

func getLevelValue(levels []dao.Level, date time.Time) int {
	avg := 0
	for _, l := range levels {
		avg += l.Level
		if util.DateEquals(time.Time(l.Date), date) {
			return l.Level
		}
	}
	return avg / len(levels)
}
