package util

import (
	"github.com/sirupsen/logrus"
	"github.com/rwcarlsen/goexif/exif"
	"image"
	"io"
	"math"
	"time"
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

func GetImageRealDate(f io.Reader) *time.Time {
	_exif, err := exif.Decode(f)
	if err != nil {
		logrus.Warn("Can't parse exif: ", err)
		return nil
	}
	dateTime, err := _exif.DateTime()
	if err != nil {
		logrus.Warn("Can't get date from exif: ", err)
		return nil
	}
	return &dateTime
}
