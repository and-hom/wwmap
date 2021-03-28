package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"image"
)

func DetectLine(src image.Image) int {
	bounds := src.Bounds()
	rightYAxisXPos := getYAxisXPos(src, bounds.Dx()-30)
	return detectBlueLineY(src, rightYAxisXPos-1)
}

func detectBlueLineY(src image.Image, xScanLine int) int {
	bounds := src.Bounds()
	for y := 60; y < bounds.Dy(); y++ {
		r, g, b, _ := src.At(xScanLine, y).RGBA()
		if isBlue(r, g, b) {
			return y
		}
	}
	return -1
}

/// Works only for 10-days scale
func DetectPrevDaysLine(src image.Image, daysOffset int) int {
	if daysOffset > 9 {
		log.Errorf("Level detection works max 9 days back. Offset is %d", daysOffset)
		return dao.NAN_LEVEL
	}
	if daysOffset < 1 {
		log.Errorf("Day offset should be 1 to 9, but was is %d", daysOffset)
		return dao.NAN_LEVEL
	}

	bounds := src.Bounds()
	rightYAxisXPos := getYAxisXPos(src, bounds.Dx()-30)

	// to the bottom end of y-axis
	y := 100
	for ; y < 300; y++ {
		r, g, b, _ := src.At(rightYAxisXPos, y).RGBA()
		if !isBlack(r, g, b) {
			break
		}
	}

	// some steps top to y-axis
	for i := 0; i < 10; i++ {
		r, g, b, _ := src.At(rightYAxisXPos-5, y-i).RGBA()
		if isBlack(r, g, b) {
			y = y - i + 1
			break
		}
	}

	// go left until marker
	labelX := rightYAxisXPos - 2
	labelNumber := 0
	for ; labelX > 0; labelX-- {
		r, g, b, _ := src.At(labelX, y).RGBA()
		if isBlack(r, g, b) {
			labelNumber += 1
		}
		if labelNumber == daysOffset {
			break
		}
	}
	if labelX == 0 {
		log.Errorf("Can't detect value for offset %d", daysOffset)
		return dao.NAN_LEVEL
	}

	return detectBlueLineY(src, labelX)
}
