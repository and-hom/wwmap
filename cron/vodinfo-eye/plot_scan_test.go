package main_test

import (
	"fmt"
	. "github.com/and-hom/wwmap/cron/vodinfo-eye"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/stretchr/testify/assert"
	"image"
	_ "image/png"
	"log"
	"os"
	"testing"
)

var patternMatcher = newTestPatternMatcher()

func newTestPatternMatcher() PatternMatcher {
	pm, err := NewPatternMatcher()
	if err != nil {
		log.Fatal("Can not initialize image pattern matcher ", err)
	}
	return pm
}

func TestNull(t *testing.T) {
	f, err := os.Open("test/null.png")
	assert.Nil(t, err)
	img, _, err := image.Decode(f)
	assert.Nil(t, err)
	l := getLevelValue(img, t)
	assert.Equal(t, dao.NAN_LEVEL, l)
}

func getLevelValue(img image.Image, t *testing.T) int {
	icd, err := Calibrate(img, &patternMatcher)
	assert.Nil(t, err)

	calibrated := CalibratedImage{
		Img:             img,
		CalibrationData: icd,
		Ok:              true,
	}

	l := calibrated.GetLevelValue(DetectLine)
	return l
}

func Test149(t *testing.T) {
	testLevelInternal(t, 149)
}

func Test460(t *testing.T) {
	testLevelInternal(t, 459)
}

func TestM73(t *testing.T) {
	testLevelInternal(t, -73)
}

func testLevelInternal(t *testing.T, value int) {
	f, err := os.Open(fmt.Sprintf("test/%d.png", value))
	assert.Nil(t, err)
	img, _, err := image.Decode(f)
	assert.Nil(t, err)
	l := getLevelValue(img, t)
	assert.Equal(t, value, l)
}
