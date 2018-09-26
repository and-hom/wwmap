package passport

import (
	"time"
	"gopkg.in/dc0d/tinykv.v4"
	"encoding/json"
	"net/http"
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"fmt"
	"errors"
)

type userInfo struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
	Id        int64 `json:"id"`
}

type Resp struct {
	Response []userInfo `json:"response"`
}

func (this userInfo) toUserInfo() UserInfo {
	login := this.Nickname
	if login == "" {
		login = fmt.Sprintf("%d", this.Id)
	}
	return UserInfo{
		Id:this.Id,
		Login:login,
		FirstName:this.FirstName,
		LastName:this.LastName,
	}
}

func Vk(cacheExpireTime time.Duration) Passport {
	kv := tinykv.New(cacheExpireTime)
	return &VkPassport{
		client:&http.Client{},
		cache:&kv,
	}
}

type VkPassport struct {
	client *http.Client
	cache  *tinykv.KV
}

func (this *VkPassport)ResolveUserInfo(token string) (UserInfo, error) {
	if info, found := (*this.cache).Get(token); found {
		return info.(UserInfo), nil
	}

	log.Info("Request user info from VK")

	url := fmt.Sprintf("https://api.vk.com/method/users.get?access_token=%s&v=5.85&fields=nickname", token)
	log.Debug(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return UserInfo{}, err
	}
	resp, err := this.client.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	if resp.StatusCode == 401 {
		return UserInfo{}, UnauthorizedError{}
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return UserInfo{}, err
	}

	vkResult := Resp{}
	err = json.Unmarshal(bytes, &vkResult)
	if err != nil {
		log.Errorf("Can not unmarshal response %s: %v", string(bytes), err)
		return UserInfo{}, err
	}
	if (len(vkResult.Response) == 0) {
		msg := fmt.Sprintf("No user entries found: %s", string(bytes))
		log.Error(msg)
		return UserInfo{}, errors.New(msg)
	}
	result := vkResult.Response[0].toUserInfo()
	(*this.cache).Put(token, result)
	return result, err
}