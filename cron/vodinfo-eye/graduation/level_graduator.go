package graduation

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"sort"
)

type LevelGraduator interface {
	Graduate([]dao.Level) ([dao.LEVEL_GRADUATION]int, error)
}

func NewPercentileGladiator(low float64, high float64) (LevelGraduator, error) {
	if low+high >= 1 {
		return nil, fmt.Errorf("Low boundary %f, high boundary %f. Overlap detected", low, high)
	}
	if low < 0 || low > 1 {
		return nil, fmt.Errorf("Invalid low boundary value %f: should be [0;1]", low)
	}
	if high < 0 || high > 1 {
		return nil, fmt.Errorf("Invalid high boundary value %f: should be [0;1]", high)
	}
	return percentileGladiator{low: low, high: high}, nil
}

type percentileGladiator struct {
	low  float64
	high float64
}

func (this percentileGladiator) Graduate(levels []dao.Level) ([dao.LEVEL_GRADUATION]int, error) {
	var result [dao.LEVEL_GRADUATION]int
	levelCnt := len(levels)

	if levelCnt == 0 {
		return result, errors.New("No levels: can't graduate")
	}

	if levelCnt < dao.LEVEL_GRADUATION {
		return result, fmt.Errorf("%d levels: can't graduate", levelCnt)
	}

	levelValues := make([]int, levelCnt)
	for i := 0; i < levelCnt; i++ {
		levelValues[i] = levels[i].Level
	}
	sort.Ints(levelValues)

	highCnt := int((1 - this.high) * float64(levelCnt))
	lowCnt := int(this.low * float64(levelCnt))

	levelValues = levelValues[lowCnt:highCnt]

	logrus.Debug(levelValues, lowCnt, highCnt)
	min := levelValues[0]
	max := levelValues[len(levelValues)-1]

	return [dao.LEVEL_GRADUATION]int{
		min,
		min + (max-min)/3,
		min + 2*(max-min)/3,
		max,
	}, nil
}
