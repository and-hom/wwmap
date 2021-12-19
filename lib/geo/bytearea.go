package geo

import (
	"fmt"
	"image"
)

type Bytearea2D interface {
	Get(x, y int) (int, error)
}

func InitBytearea2D(data [][]int) (Bytearea2D, error) {
	return bytearea2D{data}, nil
}

type bytearea2D struct {
	Data [][]int
}

func (this bytearea2D) Get(latSec, lonSec int) (int, error) {
	if latSec < 0 || latSec >= len(this.Data) {
		return 0, fmt.Errorf("Incorrect x-coord %d for area width %d", latSec, len(this.Data))
	}
	if lonSec < 0 || lonSec >= len(this.Data[latSec]) {
		return 0, fmt.Errorf("Incorrect coords %d %d for area %dx%d", latSec, lonSec, len(this.Data), len(this.Data[latSec]))
	}
	return this.Data[3600-latSec][lonSec], nil
}

func InitImageBasedBytearea2D(img image.Image) (Bytearea2D, error) {
	return &imageBasedByteArea{img}, nil
}

type imageBasedByteArea struct {
	Image image.Image
}

func (this *imageBasedByteArea) Get(latSec, lonSec int) (int, error) {
	r, _, _, _ := this.Image.At(lonSec, 3600-latSec).RGBA()
	return int(r), nil
}
