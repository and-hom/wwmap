package passport

import (
	"net/http"
	"gopkg.in/dc0d/tinykv.v4"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	"time"
	"io/ioutil"
)

type Sex string

const Male Sex = "male"
const Female Sex = "female"

type UserInfo struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	DisplayName     string `json:"display_name"`
	RealName        string `json:"real_name"`
	IsAvatarEmpty   bool `json:"is_avatar_empty"`
	DefaultAvatarId string `json:"default_avatar_id"`
	Login           string `json:"login"`
	Sex             Sex `json:"sex"`
	Id              int64 `json:"id,string"`
}

func New(cacheExpireTime time.Duration) YandexPassport {
	kv := tinykv.New(cacheExpireTime)
	return YandexPassport{
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
		(*this.cache).Put(token, result)
	}

	return result, err
}

type UnauthorizedError struct {

}

func (this UnauthorizedError) Error() string {
	return "Unauthorized"
}