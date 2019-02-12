package main

import "image"

func DetectLine(src image.Image) int {
	bounds := src.Bounds()
	rightYAxisXPos := getYAxisXPos(src, bounds.Dx()-30)
	return detectBlueLineY(src, rightYAxisXPos-1)
}

func detectBlueLineY(src image.Image, xScanLine int) int {
	bounds := src.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		r, g, b, _ := src.At(xScanLine, y).RGBA()
		if isBlue(r, g, b) {
			return y
		}
	}
	return -1
}

/// Works only for 10-days scale
func DetectYesterdayLine(src image.Image) int {
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

	for i := 0; i < 10; i++ {
		r, g, b, _ := src.At(rightYAxisXPos-5, y-i).RGBA()
		if isBlack(r, g, b) {
			y = y - i + 1
			break
		}
	}

	labelX := rightYAxisXPos - 2
	for ; labelX > 0; labelX-- {
		r, g, b, _ := src.At(labelX, y).RGBA()
		if isBlack(r, g, b) {
			break
		}
	}

	return detectBlueLineY(src, labelX)
}
