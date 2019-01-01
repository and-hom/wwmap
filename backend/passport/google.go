package passport

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/dc0d/tinykv.v4"
	"io/ioutil"
	"net/http"
	"time"
)

func Google(cacheExpireTime time.Duration) Passport {
	kv := tinykv.New(cacheExpireTime)
	return &GooglePassport{
		client: &http.Client{},
		cache:  &kv,
	}
}

type GoogleUserInfo struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
}

func (this GoogleUserInfo) toUserInfo() UserInfo {
	return UserInfo{
		FirstName:       this.GivenName,
		LastName:        this.FamilyName,
		Login:           this.Email,
		Id:              this.Id,
		Sex:             Sex(this.Gender),
		DefaultAvatarId: this.Picture,
		DisplayName:     this.Name,
	}
}

type GooglePassport struct {
	client *http.Client
	cache  *tinykv.KV
}

func (this *GooglePassport) ResolveUserInfo(token string) (UserInfo, error) {
	if info, found := (*this.cache).Get(token); found {
		return info.(UserInfo), nil
	}

	log.Info("Request user info from Google")

	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v1/userinfo?access_token=%s", token)
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

	googleResult := GoogleUserInfo{}
	err = json.Unmarshal(bytes, &googleResult)
	if err != nil {
		log.Errorf("Can not unmarshal response %s: %v", string(bytes), err)
		return UserInfo{}, err
	}
	result := googleResult.toUserInfo()
	(*this.cache).Put(token, result)
	return result, err
}
