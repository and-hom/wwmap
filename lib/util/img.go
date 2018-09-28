package util

import (
	"image"
	"math"
)

func PreviewRect(r image.Rectangle, areaWidth, areaHeight int) (image.Rectangle, bool) {
	srcWidth := r.Max.X - r.Min.X
	srcHeight := r.Max.Y - r.Min.Y

	kX := float64(areaWidth) / float64(srcWidth)
	kY := float64(areaHeight) / float64(srcHeight)

	newImgWidht := areaWidth
	newImgHeight := areaHeight

	k := math.Min(kX, kY)
	newImgWidht = int(k * float64(srcWidth))
	newImgHeight = int(k * float64(srcHeight))

	return image.Rect(0, 0, newImgWidht, newImgHeight), kX > 1.0 && kY > 1.0
}