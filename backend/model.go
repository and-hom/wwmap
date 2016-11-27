package main

import (
	"bytes"
	"fmt"
)

type Point struct {
	x float64
	y float64
}

func (this Point) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[")
	buffer.WriteString(fmt.Sprint(this.x))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprint(this.y))
	buffer.WriteString("]")
	return buffer.Bytes(), nil
}

type Track struct {
	Title string `json:"title"`
	Path  []Point `json:"path"`
}

type ExtDataTrack struct {
	Title string `json:"title"`
	FileIds   []string `json:"fileIds"`
}
