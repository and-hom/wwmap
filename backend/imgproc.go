package main

import (
	"io"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"time"
	"fmt"
	. "github.com/and-hom/wwmap/lib/geo"
)

type ImgProperties struct {
	Coords    *Point `json:"coords,omitempty"`
	Timestamp *ImgPropertiesFormattedTime `json:"timestamp,omitempty"`
}

type ImgPropertiesFormattedTime time.Time

func (t ImgPropertiesFormattedTime)MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05 MST"))
	return []byte(stamp), nil
}

func GetImgProperties(file io.Reader) (*ImgProperties, error) {
	exif.RegisterParsers(mknote.All...)

	_exif, err := exif.Decode(file)
	if err != nil {
		return nil, err
	}

	imgData := ImgProperties{}
	addTime(_exif, &imgData)
	addCoords(_exif, &imgData)
	return &imgData, nil
}

func addCoords(_exif *exif.Exif, imgData *ImgProperties) {
	lat, lon, err := _exif.LatLong()
	if err == nil {
		point := Point{
			Lat:lat,
			Lon:lon,
		}
		imgData.Coords = &point
	}
}

func addTime(_exif *exif.Exif, imgData *ImgProperties) {
	time, err := _exif.DateTime()
	if err == nil {
		jsonTime := ImgPropertiesFormattedTime(time)
		imgData.Timestamp = &jsonTime
	}
}
