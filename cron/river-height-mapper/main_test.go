package main_test

import (
	rhm "github.com/and-hom/wwmap/cron/river-height-mapper"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestGetVectorNormPointStep0(t *testing.T) {
	dLat, dLon := rhm.GetVectorNormPoint(0, 1)
	assert.Equal(t, 0.0, dLat)
	assert.Equal(t, 0.0, dLon)
}

func TestGetVectorNormPoint45(t *testing.T) {
	dLat, dLon := rhm.GetVectorNormPoint(1, 1)
	assert.InDelta(t, -1/math.Sqrt(2), dLat, 0.001)
	assert.InDelta(t, 1/math.Sqrt(2), dLon, 0.001)
}

func TestGetVectorNormPoint0(t *testing.T) {
	dLat, dLon := rhm.GetVectorNormPoint(1, 0)
	assert.InDelta(t, -1.0, dLat, 0.001)
	assert.InDelta(t, 0.0, dLon, 0.001)
}

func TestGetVectorNormPoint90(t *testing.T) {
	dLat, dLon := rhm.GetVectorNormPoint(1, math.Inf(1))
	assert.InDelta(t, 0.0, dLat, 0.001)
	assert.InDelta(t, 1.0, dLon, 0.001)
}

func TestGetVectorNormPointMinus90(t *testing.T) {
	dLat, dLon := rhm.GetVectorNormPoint(1, math.Inf(-1))
	assert.InDelta(t, 0.0, dLat, 0.001)
	assert.InDelta(t, -1.0, dLon, 0.001)
}

func TestGetVectorNormPoint30(t *testing.T) {
	dLat, dLon := rhm.GetVectorNormPoint(1, 3.0/4.0)
	assert.InDelta(t, -4.0/5.0, dLat, 0.001)
	assert.InDelta(t, 3.0/5.0, dLon, 0.001)
}

