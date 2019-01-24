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
