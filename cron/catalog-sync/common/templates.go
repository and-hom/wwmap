package common

import (
	"html/template"
	"fmt"
	"math"
	"bytes"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/and-hom/wwmap/lib/model"
	"strconv"
)

type Templates interface {
	WriteSpot(data interface{}) (string, error)
	WriteRiver(data interface{}) (string, error)
	WriteRegion(data interface{}) (string, error)
	WriteCountry(data interface{}) (string, error)
	WriteRoot(data interface{}) (string, error)
}

type templates struct {
	Spot      *template.Template
	River     *template.Template
	Region    *template.Template
	Country   *template.Template
	Root      *template.Template
	Decorator *template.Template
}

type DecoratorParams struct {
	Body template.HTML
	Data interface{}
}

func (this *templates) WriteSpot(data interface{}) (string, error) {
	return this.WithDecorator(this.Spot, data)
}

func (this *templates) WriteRiver(data interface{}) (string, error) {
	return this.WithDecorator(this.River, data)
}

func (this *templates) WriteRegion(data interface{}) (string, error) {
	return this.WithDecorator(this.Region, data)
}

func (this *templates) WriteCountry(data interface{}) (string, error) {
	return this.WithDecorator(this.Country, data)
}

func (this *templates) WriteRoot(data interface{}) (string, error) {
	return this.WithDecorator(this.Root, data)
}

func (this *templates) WithDecorator(t *template.Template, data interface{}) (string, error) {
	internalBuf := bytes.Buffer{}
	err := t.Execute(&internalBuf, data)
	if err != nil {
		log.Errorf("Can not process river template", err)
		return "", err
	}

	fullBuf := bytes.Buffer{}
	err = this.Decorator.Execute(&fullBuf, DecoratorParams{Body:template.HTML(internalBuf.String()), Data:data, })
	if err != nil {
		log.Errorf("Can not process river template", err)
		return "", err
	}

	return fullBuf.String(), nil
}

func LoadTemplates(load func(name string) []byte) (Templates, error) {
	const nbsp = '\u00A0'
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"spotCatStr": CategoryStr,
		"catStr": func(cat model.SportCategory) string {
			return util.HumanReadableCategoryName(cat, false)
		},
		"ccol": func(cat model.SportCategory) string {
			switch cat.Category {
			case 1:
				return "light-blue"
			case 2:
				return "green"
			case 3:
				return "yellow"
			case 4:
				return "orange"
			case 5:
				return "red"
			case 6:
				return "#990000"
			default:
				return "dark-grey"
			}
		},
		"coalesce_string_prop": func(name string, _default string, props ...map[string]interface{}) string {
			for _, p := range props {
				foundVal, found := p[name]
				if found && foundVal != nil {
					strVal, castOk := foundVal.(string)
					if castOk {
						return strVal
					}
				}
			}
			return _default
		},
		"coalesce_int_prop": func(name string, _default int, props ...map[string]interface{}) int {
			for _, p := range props {
				foundVal, found := p[name]
				if found && foundVal != nil {
					intVal, castOk := foundVal.(int64)
					if castOk {
						return int(intVal)
					}
					floatVal, castOk := foundVal.(float64)
					if castOk {
						return int(floatVal)
					}
					intValS, castOk := foundVal.(string)
					if castOk {
						intVal, err := strconv.ParseInt(intValS, 10, 32)
						if err != nil {
							return int(intVal)
						}
					}
				}
			}
			return _default
		},
		"lat":func(lat float64) string {
			var abbr string
			if lat > 0 {
				abbr = "с.ш."
			} else {
				abbr = "ю.ш."
			}
			return fmt.Sprintf("%.7f%c%s", math.Abs(lat),nbsp, abbr)
		},
		"lon":func(lon float64) string {
			var abbr string
			if lon > 0 {
				abbr = "в.д."
			} else {
				abbr = "з.д."
			}
			return fmt.Sprintf("%.7f%c%s", math.Abs(lon),nbsp, abbr)
		},
	}
	t := templates{}
	var e error

	t.Spot, e = template.New("spot").Funcs(funcMap).Parse(string(load("spot-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile spot template: %s", e.Error())
	}
	t.River, e = template.New("river").Funcs(funcMap).Parse(string(load("river-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile river template: %s", e.Error())
	}
	t.Region, e = template.New("region").Funcs(funcMap).Parse(string(load("region-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile region template: %s", e.Error())
	}
	t.Country, e = template.New("country").Funcs(funcMap).Parse(string(load("country-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile country template: %s", e.Error())
	}
	t.Root, e = template.New("root").Funcs(funcMap).Parse(string(load("root-page-template.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile root template: %s", e.Error())
	}
	t.Decorator, e = template.New("decorator").Funcs(funcMap).Parse(string(load("decorator.htm")))
	if e != nil {
		return nil, fmt.Errorf("Can not compile decorator template: %s", e.Error())
	}
	return &t, nil
}

func CategoryStr(spot dao.WhiteWaterPointFull) string {
	if (!spot.HighWaterCategory.Undefined() || !spot.MediumWaterCategory.Undefined() || !spot.LowWaterCategory.Undefined()) {
		return fmt.Sprintf("%s/%s/%s",
			util.HumanReadableCategoryName(spot.LowWaterCategory, false),
			util.HumanReadableCategoryName(spot.MediumWaterCategory, false),
			util.HumanReadableCategoryName(spot.HighWaterCategory, false), )
	}
	return util.HumanReadableCategoryName(spot.Category, false)
}