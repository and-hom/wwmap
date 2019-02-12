package main

import "fmt"

func InitialImageCalibrationData() ImageCalibrationData {
	return ImageCalibrationData{
		LMin: 100000.0,
		YMin: 100000.0,
		LMax: -100000.0,
		YMax: -100000.0,
	}
}

type ImageCalibrationData struct {
	LMin, YMin, LMax, YMax float64
}

func (this *ImageCalibrationData) Add(l, y float64) {
	if l > this.LMax {
		this.LMax = l
		this.YMax = y
	}
	if l < this.LMin {
		this.LMin = l
		this.YMin = y
	}
}

func (this *ImageCalibrationData) YToLevel(y float64) float64 {
	return (this.YMin-y)/(this.YMin-this.YMax)*(this.LMax-this.LMin) + this.LMin
}

func (this *ImageCalibrationData) String() string {
	return fmt.Sprintf("LMin=%f YMin=%f LMax=%f YMax=%f", this.LMin, this.YMin, this.LMax, this.YMax)
}
