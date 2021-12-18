package geo

import "fmt"

type Bytearea2D interface {
	Get(x, y int) (int32, error)
}

func InitBytearea2D(data [][]int32) (Bytearea2D, error) {
	return bytearea2D{data}, nil
}

type bytearea2D struct {
	Data   [][]int32
}

func (this bytearea2D) Get(x, y int) (int32, error) {
	if x < 0 || x >= len(this.Data) {
		return 0, fmt.Errorf("Incorrect x-coord %d for area width %d", x, len(this.Data))
	}
	if y < 0 || y >= len(this.Data[x]) {
		return 0, fmt.Errorf("Incorrect coords %d %d for area %dx%d", x, y, len(this.Data), len(this.Data[x]))
	}
	return this.Data[x][y], nil
}
