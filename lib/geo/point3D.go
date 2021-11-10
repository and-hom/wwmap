package geo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
)

type Point3D struct {
	Lat float64
	Lon float64
	Alt int64
}

func (this Point3D) Flip() Point3D {
	return Point3D{
		Lat: this.Lon,
		Lon: this.Lat,
		Alt: this.Alt,
	}
}

func (this Point3D) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("[")
	buffer.WriteString(fmt.Sprint(this.Lat))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprint(this.Lon))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprint(this.Alt))
	buffer.WriteString("]")
	return buffer.Bytes(), nil
}

func (this *Point3D) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	arr := make([]json.Number, 3)
	err := decoder.Decode(&arr)
	if err != nil {
		return err
	}
	this.Lat, err = arr[0].Float64()
	if err != nil {
		return err
	}
	this.Lon, err = arr[1].Float64()
	if err != nil {
		return err
	}
	this.Alt, err = arr[2].Int64()
	if err != nil {
		return err
	}
	return nil
}

func (this Point3D) DistanceTo(p Point3D) float64 {
	return math.Sqrt(
		(this.Lat-p.Lat)*(this.Lat-p.Lat) +
			(this.Lon-p.Lon)*(this.Lon-p.Lon) +
			float64((this.Alt-p.Alt)*(this.Alt-p.Alt)),
	)
}

func (this *Point3D) String() string {
	return fmt.Sprintf("(lat=%f, lon=%f, alt=%d)", this.Lat, this.Lon, this.Alt)
}
