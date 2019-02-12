package main_test

import (
	. "github.com/and-hom/wwmap/cron/vodinfo-eye"
	"github.com/stretchr/testify/assert"
	"image"
	_ "image/png"
	"os"
	"testing"
)

func TestDetectYesterdayLine(t *testing.T) {
	f, err := os.Open("test/yesterday-data.png")
	assert.Nil(t, err)
	img, _, err := image.Decode(f)
	assert.Nil(t, err)
	yLine := DetectYesterdayLine(img)
	assert.Equal(t, 160, yLine)
}

func TestDetectYesterdayLine2(t *testing.T) {
	f, err := os.Open("test/-73.png")
	assert.Nil(t, err)
	img, _, err := image.Decode(f)
	assert.Nil(t, err)
	yLine := DetectYesterdayLine(img)
	assert.Equal(t, 148, yLine)
}

func TestDetectYesterdayLineNull(t *testing.T) {
	f, err := os.Open("test/null.png")
	assert.Nil(t, err)
	img, _, err := image.Decode(f)
	assert.Nil(t, err)
	yLine := DetectYesterdayLine(img)
	assert.Equal(t, -1, yLine)
}
