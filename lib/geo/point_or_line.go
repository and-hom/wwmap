package geo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type PointOrLine struct {
	Point *Point   `json:"point,omitempty"`
	Line  *[]Point `json:"line,omitempty"`
}

func (this PointOrLine) Center() Point {
	if this.Point != nil {
		return *this.Point
	}
	if this.Line != nil {
		return (*this.Line)[0] // todo real center ?
	}
	return Point{}
}

func (this PointOrLine) MarshalJSON() ([]byte, error) {
	if this.Point != nil {
		return json.Marshal(this.Point)
	}
	if this.Line != nil {
		return json.Marshal(this.Line)
	}
	return []byte{}, errors.New("Can not serialize empty geo object")
}

func (this *PointOrLine) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	t, err := d.Token()
	if err != nil {
		return err
	}
	if t == json.Delim('[') {
		t2, err := d.Token()
		if err != nil {
			return err
		}
		if t2 == json.Delim('[') {
			if err := json.Unmarshal(data, &this.Line); err != nil {
				return err
			}
		} else {
			this.Point = &Point{}
			if err := json.Unmarshal(data, this.Point); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("Expected [ but found %v", t)
}

func (this PointOrLine) ToPg() interface{} {
	if this.Point != nil {
		return NewPgGeoPoint(*this.Point)
	} else if this.Line != nil {
		return NewPgLineString(*this.Line)
	}
	return nil
}

func (this PointOrLine) Flip() PointOrLine {
	var filppedPointPtr *Point
	if this.Point != nil {
		flipped := (*this.Point).Flip()
		filppedPointPtr = &flipped
	}

	var flippedLine *[]Point
	if this.Line != nil {
		fl := make([]Point, len(*this.Line))
		for i := 0; i < len(*this.Line); i++ {
			fl[i] = (*this.Line)[i].Flip()
		}
		flippedLine = &fl
	}

	return PointOrLine{
		Point: filppedPointPtr,
		Line:  flippedLine,
	}
}
