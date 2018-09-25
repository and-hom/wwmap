package common

import (
	"html/template"
	"fmt"
	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/util"
	"github.com/and-hom/wwmap/lib/model"
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
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"spotCatStr": CategoryStr,
		"catStr": func(cat model.SportCategory) string {
			return util.HumanReadableCategoryName(cat, false)
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