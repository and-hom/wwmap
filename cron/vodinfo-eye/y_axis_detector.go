package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"image"
)

func DetectYAxisLabels(src image.Image, yAxisLabelsCoords map[int]int) (map[int]int, error) {
	yAxisXPos := getYAxisXPos(src, 0)
	if yAxisXPos < 0 {
		return nil, fmt.Errorf("Y axis is not detected!")
	}
	log.Debugf("Y axis detected at %d", yAxisXPos)

	yAxisMarks := getYAxisMarks(src, yAxisXPos)
	log.Debugf("Y axis marks are %v", yAxisMarks)

	result := make(map[int]int)
	for val, labelY := range yAxisLabelsCoords {
		for _, markY := range yAxisMarks {
			if markY > labelY-7 && markY < labelY+17 {
				result[val] = markY
			}
		}
	}

	return result, nil
}

func getYAxisXPos(src image.Image, xOffset int) int {
	bounds := src.Bounds()
	yScan := bounds.Dy() / 2
	for x := xOffset; x < bounds.Dx(); x++ {
		detected := true
		for y := -20; y < 20; y++ {
			r, g, b, _ := src.At(x, yScan+y).RGBA()

			if !isBlack(r, g, b) {
				detected = false
				break
			}
		}
		if detected {
			return x
		}
	}
	return -1
}

func getYAxisMarks(src image.Image, yAxisXPos int) []int {
	marksXPos := yAxisXPos - 2
	bounds := src.Bounds()
	result := []int{0}
	blackCounter := 0
	for y := 0; y < bounds.Dy(); y++ {
		r, g, b, _ := src.At(marksXPos, y).RGBA()
		if isBlack(r, g, b) {
			blackCounter++
		} else {
			if blackCounter > 0 && blackCounter < 6 {
				result = append(result, y-blackCounter/2)
			}
			blackCounter = 0
		}
	}
	return result
}
