package main_test

import (
	. "github.com/and-hom/wwmap/cron/vodinfo-eye"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitial(t *testing.T) {
	d := InitialImageCalibrationData()
	d.Add(1.0, 2.0)
	assert.Equal(t, ImageCalibrationData{1.0, 2.0, 1.0, 2.0}, d)
}

func TestAdd(t *testing.T) {
	d := ImageCalibrationData{
		LMax: 0,
		YMax: 0,
		LMin: 0,
		YMin: 0,
	}

	d.Add(10, -100)
	assert.Equal(t, 0.0, d.LMin)
	assert.Equal(t, 0.0, d.YMin)
	assert.Equal(t, 10.0, d.LMax)
	assert.Equal(t, -100.0, d.YMax)

	d.Add(8, -120)
	assert.Equal(t, 0.0, d.LMin)
	assert.Equal(t, 0.0, d.YMin)
	assert.Equal(t, 10.0, d.LMax)
	assert.Equal(t, -100.0, d.YMax)

	d.Add(-2, -220)
	assert.Equal(t, -2.0, d.LMin)
	assert.Equal(t, -220.0, d.YMin)
	assert.Equal(t, 10.0, d.LMax)
	assert.Equal(t, -100.0, d.YMax)
}

func TestY2Level(t *testing.T) {
	d := ImageCalibrationData{
		LMax: 10,
		YMax: 100,
		LMin: 0,
		YMin: 0,
	}
	assert.Equal(t, 5.0, d.YToLevel(50.0))
	assert.Equal(t, 15.0, d.YToLevel(150.0))
	assert.Equal(t, -5.0, d.YToLevel(-50.0))
}
