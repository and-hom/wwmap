package main_test

import (
	"github.com/and-hom/wwmap/cron/vodinfo-eye/graduation"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/stretchr/testify/assert"
	_ "image/png"
	"testing"
)

var sampleLevels = []dao.Level{
	{Level: 00},
	{Level: 10},
	{Level: 20},
	{Level: 30},
	{Level: 40},
	{Level: 50},
	{Level: 60},
	{Level: 70},
	{Level: 80},
	{Level: 90},
	{Level: 100},
	{Level: 110},
	{Level: 120},
}

func TestPercentileGladiator_Graduate_0vals(t *testing.T) {
	graduator, err := graduation.NewPercentileGladiator(0, 0)
	assert.Nil(t, err)
	_, err = graduator.Graduate([]dao.Level{})
	assert.NotNil(t, err)
}

func TestPercentileGladiator_Graduate_Overlap(t *testing.T) {
	_, err := graduation.NewPercentileGladiator(0, 1)
	assert.NotNil(t, err)
}

func TestPercentileGladiator_Graduate_0_0(t *testing.T) {
	graduator, err := graduation.NewPercentileGladiator(0, 0)
	assert.Nil(t, err)
	assert.NotNil(t, graduator)
	graduation, err := graduator.Graduate(sampleLevels)
	assert.Nil(t, err)
	assert.Equal(t, 00, graduation[0])
	assert.Equal(t, 40, graduation[1])
	assert.Equal(t, 80, graduation[2])
	assert.Equal(t, 120, graduation[3])
}

func TestPercentileGladiator_Graduate_0_03(t *testing.T) {
	graduator, err := graduation.NewPercentileGladiator(0, 0.3)
	assert.Nil(t, err)
	assert.NotNil(t, graduator)
	graduation, err := graduator.Graduate(sampleLevels)
	assert.Nil(t, err)
	assert.Equal(t, 00, graduation[0])
	assert.Equal(t, 26, graduation[1])
	assert.Equal(t, 53, graduation[2])
	assert.Equal(t, 80, graduation[3])
}

func TestPercentileGladiator_Graduate_03_0(t *testing.T) {
	graduator, err := graduation.NewPercentileGladiator(0.34, 0)
	assert.Nil(t, err)
	assert.NotNil(t, graduator)
	graduation, err := graduator.Graduate(sampleLevels)
	assert.Nil(t, err)
	assert.Equal(t, 40, graduation[0])
	assert.Equal(t, 66, graduation[1])
	assert.Equal(t, 93, graduation[2])
	assert.Equal(t, 120, graduation[3])
}

func TestPercentileGladiator_Graduate_025_025(t *testing.T) {
	graduator, err := graduation.NewPercentileGladiator(0.25, 0.25)
	assert.Nil(t, err)
	assert.NotNil(t, graduator)
	graduation, err := graduator.Graduate(sampleLevels)
	assert.Nil(t, err)
	assert.Equal(t, 30, graduation[0])
	assert.Equal(t, 46, graduation[1])
	assert.Equal(t, 63, graduation[2])
	assert.Equal(t, 80, graduation[3])
}
