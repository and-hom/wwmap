package passport

import (
	"time"
	"gopkg.in/dc0d/tinykv.v4"
	"encoding/json"
	"net/http"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
)

func Yandex(cacheExpireTime time.Duration) Passport {
	kv := tinykv.New(cacheExpireTime)
	return &YandexPassport{
		client:&http.Client{},
		cache:&kv,
	}
}

type YandexPassport struct {
	client *http.Client
	cache  *tinykv.KV
}

func (this *YandexPassport)ResolveUserInfo(token string) (UserInfo, error) {
	if info, found := (*this.cache).Get(token); found {
		return info.(UserInfo), nil
	}

	log.Info("Request user info from Yandex Passport")
	result := UserInfo{}

	req, err := http.NewRequest("GET", "https://login.yandex.ru/info?format=json", nil)
	if err != nil {
		return result, err
	}
	req.Header.Add("Authorization", "OAuth " + token)
	resp, err := this.client.Do(req)
	if err != nil {
		return result, err
	}
	if resp.StatusCode==401 {
		return result, UnauthorizedError{}
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		log.Errorf("Can not unmarshal response %s: %v", string(bytes), err)
	} else {
		(*this.cache).Put(token, result)
	}
	return result, err
}