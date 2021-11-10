package geo_test

import (
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)
const FLOAT_COMPARSION_TOLERANCE = 0.001

func TestPoint3DUnmarshal(t *testing.T) {
	p := geo.Point3D{}
	err := p.UnmarshalJSON([]byte("[45.43, 14.12, 100]"))
	assert.Nil(t, err)
	assert.Equal(t, geo.Point3D{45.43, 14.12, 100}, p)
}

func TestPoint3DMarshal(t *testing.T) {
	p := geo.Point3D{45.43, 14.12, 100}
	json, err := p.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, []byte("[45.43,14.12,100]"), json)
}

func TestPoint3DDistance(t *testing.T) {
	p1 := geo.Point3D{2, 3, 4}
	p2 := geo.Point3D{0, 0, 0}
	d := p1.DistanceTo(p2)
	assert.True(t, math.Abs(5.385 - d) < FLOAT_COMPARSION_TOLERANCE)
}
