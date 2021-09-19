package main

import (
	"github.com/and-hom/wwmap/lib/util"
	"math"
)

func RemoveHeightGrowing(heights []int32) {
	start := int32(0)
	end := int32(0)
	for i := 0; i < util.Min(10, len(heights)); i++ {
		start += heights[i]
	}
	for i := len(heights) - 1; i >= util.Max(0, len(heights)-10); i-- {
		end += heights[i]
	}
	reverse := start < end

	if (reverse) {
		return
	}

	if (reverse) {
		for i := 0; i < len(heights); i++ {
			heights[i] = math.MaxInt16 - heights[i]
		}
	}

	prev := heights[0]
	interpolationStart := -1
	for i := 1; i < len(heights); i++ {
		val := heights[i]
		if val > prev {
			if interpolationStart < 0 {
				interpolationStart = i
			}
		} else {
			if interpolationStart > 0 {
				for j := interpolationStart; j < i; j++ {
					heights[j] = int32(float64(prev) - float64(prev-val)*float64(j-interpolationStart+1)/float64(i-interpolationStart+1))
				}
				interpolationStart = -1
			}
			prev = val
		}
	}

	if (reverse) {
		for i := 0; i < len(heights); i++ {
			heights[i] = math.MaxInt16 - heights[i]
		}
	}
}
