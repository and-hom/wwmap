package util_test

import (
	"github.com/and-hom/wwmap/lib/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotTrim(t *testing.T) {
	source := "abcd"
	result := util.TrimToLengthWithTrailingDots(source, 1000)
	assert.Equal(t, source, result)
}

func TestTrimEq(t *testing.T) {
	source := "abcd"
	result := util.TrimToLengthWithTrailingDots(source, 4)
	assert.Equal(t, source, result)
}

func TestTrim(t *testing.T) {
	result := util.TrimToLengthWithTrailingDots("abcdef", 4)
	assert.Equal(t, "a...", result)
}


func TestTrimShortTail0(t *testing.T) {
	result := util.TrimToLengthWithTrailingDots("abcdef", 0)
	assert.Equal(t, "", result)
}


func TestTrimShortTail1(t *testing.T) {
	result := util.TrimToLengthWithTrailingDots("abcdef", 1)
	assert.Equal(t, "a", result)
}

func TestTrimShortTail2(t *testing.T) {
	result := util.TrimToLengthWithTrailingDots("abcdef", 2)
	assert.Equal(t, "ab", result)
}

func TestTrimShortTail3(t *testing.T) {
	result := util.TrimToLengthWithTrailingDots("abcdef", 3)
	assert.Equal(t, "a..", result)
}
