package geo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
)

type Point struct {
	Lat float64
	Lon float64
}

func (this Point) Flip() Point {
	return Point{
		Lat: this.Lon,
		Lon: this.Lat,
	}
}

func (this Point) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[")
	buffer.WriteString(fmt.Sprint(this.Lat))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprint(this.Lon))
	buffer.WriteString("]")
	return buffer.Bytes(), nil
}

func (this *Point) UnmarshalJSON(data []byte) error {
	arr := make([]float64, 2)
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}
	this.Lat = arr[0]
	this.Lon = arr[1]
	return nil
}

func (this Point) DistanceTo(p Point) float64 {
	return math.Sqrt((this.Lat-p.Lat)*(this.Lat-p.Lat) + (this.Lon-p.Lon)*(this.Lon-p.Lon))
}

func (this *Point) String() string {
	return fmt.Sprintf("(lat=%f, lon=%f)", this.Lat, this.Lon)
}
