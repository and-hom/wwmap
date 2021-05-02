package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"net/http"
	"strings"
	"time"
)

type WeatherApi interface {
	Get(point geo.Point) ([]dao.Meteo, error)
}

func NewYandexWeatherApi(token string) WeatherApi {
	return &yandexWeatherApi{
		client: &http.Client{
			Timeout: 4 * time.Second,
		},
		token: token,
	}
}

type yandexWeatherApi struct {
	client *http.Client
	token  string
}

func (this yandexWeatherApi) Get(point geo.Point) ([]dao.Meteo, error) {
	log.Info("Fetch weather for ", point.String())

	url := fmt.Sprintf("https://api.weather.yandex.ru/v1/informers?lat=%f&lon=%f", point.Lat, point.Lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []dao.Meteo{}, err
	}
	req.Header.Set("X-Yandex-API-Key", this.token)

	resp, err := this.client.Do(req)
	if err != nil {
		return []dao.Meteo{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return []dao.Meteo{}, fmt.Errorf("HTTP code is %d: %s", resp.StatusCode, resp.Status)
	}

	ywr := YandexWeatherResponse{}
	err = json.NewDecoder(resp.Body).Decode(&ywr)
	if err != nil {
		return []dao.Meteo{}, err
	}

	return []dao.Meteo{
		{Temp: ywr.Fact.Temp, Rain: this.rainLevel(ywr.Fact.Condition), Daytime: this.dayTime(ywr.Fact.Daytime), Date: dao.JSONDate(ywr.Now)},
	}, nil
}

func (this yandexWeatherApi) dayTime(daytime string) dao.Daytime {
	switch daytime {
	case "n":
		return dao.NIGHT
	case "d":
		return dao.DAY
	}
	log.Error("Unknown daytime ", daytime, ". Use Night")
	return dao.NIGHT
}

func (this yandexWeatherApi) rainLevel(condition string) int {
	if strings.Contains(condition, "light-rain") {
		return 1
	}
	if strings.Contains(condition, "rain") {
		return 2
	}
	return 0
}

type YandexWeatherResponse struct {
	Fact YandexWeatherFact `json:"fact"`
	Now  dao.JSONUnixTime  `json:"now"`
}

type YandexWeatherFact struct {
	Temp       int    `json:"temp"`
	Condition  string `json:"condition"`
	PressureMm int    `json:"pressure_mm"`
	PressurePa int    `json:"pressure_pa"`
	Daytime    string `json:"daytime"`
}
