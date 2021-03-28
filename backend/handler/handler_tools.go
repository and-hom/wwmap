package handler

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/backend/passport"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/geo"
	"github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func (this *App) bboxFormValue(w http.ResponseWriter, req *http.Request) (geo.Bbox, error) {
	bboxStr := req.FormValue("bbox")
	bbox, err := geo.ParseBbox(bboxStr)
	if err != nil {
		OnError(w, err, fmt.Sprintf("Can not parse bbox: %v", bbox), http.StatusBadRequest)
		return geo.Bbox{}, err
	}
	return bbox, nil
}

func (this *App) tileParams(w http.ResponseWriter, req *http.Request) (string, geo.Bbox, error) {
	callback := req.FormValue("callback")
	bbox, err := this.bboxFormValue(w, req)
	if err != nil {
		return "", geo.Bbox{}, err
	}

	return callback, bbox, nil
}

func (this *App) tileParamsZ(w http.ResponseWriter, req *http.Request) (string, geo.Bbox, int, error) {

	callback := req.FormValue("callback")
	bbox, err := this.bboxFormValue(w, req)
	if err != nil {
		return "", geo.Bbox{}, 0, err
	}

	zoomStr := req.FormValue("zoom")
	zoom, err := strconv.Atoi(zoomStr)
	if err != nil {
		OnError(w, err, fmt.Sprintf("Can not parse zoom value: %s", zoomStr), http.StatusBadRequest)
		return "", geo.Bbox{}, 0, err
	}

	return callback, bbox, zoom, nil
}

func (this *App) CloseAndRemove(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}

func (this *App) CreateMissingUser(r *http.Request, authProvider dao.AuthProvider, info passport.UserInfo, sessionId string) (int64, dao.Role, string, bool, error) {
	return this.UserDao.CreateIfNotExists(dao.User{
		ExtId:        info.Id,
		AuthProvider: authProvider,
		Role:         dao.USER,
		Info: dao.UserInfo{
			FirstName: info.FirstName,
			LastName:  info.LastName,
			Login:     info.Login,
		},
		SessionId: sessionId,
	})
}

func (this *App) ForRoles(payload handler.HandlerFunction, roles ...dao.Role) handler.HandlerFunction {
	return handler.ForRoles(payload, this.UserDao, roles...)
}

func (this *App) GetUserInfo(r *http.Request) (dao.AuthProvider, passport.UserInfo, error) {
	providerAndToken := GetOauthProviderAndToken(r)
	p, found := this.AuthProviders[providerAndToken.AuthProvider]
	if !found {
		return dao.AuthProvider(""), passport.UserInfo{}, fmt.Errorf("Can not find provider for %v", providerAndToken.AuthProvider)
	}
	userInfo, err := p.ResolveUserInfo(providerAndToken.Token)
	return providerAndToken.AuthProvider, userInfo, err
}

func (this *App) collectReferer(r *http.Request) {
	referer := r.Header.Get("Referer")
	if referer == "" {
		return
	}

	refererUrl, err := url.Parse(referer)
	if err != nil {
		log.Warnf("Invalid referer: %s", referer)
		return
	}
	if refererUrl.Hostname() == "localhost" ||
		refererUrl.Hostname() == "wwmap.ru" ||
		strings.Contains(refererUrl.Hostname(), "yandex") ||
		strings.Contains(refererUrl.Hostname(), "vk.com") {
		return
	}

	err = this.RefererStorage.Put(refererUrl)
	if err != nil {
		log.Error("Can not store referer ", err)
	}
}

const VOYAGE_REPORT_LOG_ENTRY_TYPE = "REPORT"
const CAMP_LOG_ENTRY_TYPE = "CAMP"
const TRANSFER_LOG_ENTRY_TYPE = "TRANSFER"
const SPOT_LOG_ENTRY_TYPE = "SPOT"
const RIVER_LOG_ENTRY_TYPE = "RIVER"
const REGION_LOG_ENTRY_TYPE = "REGION"
const IMAGE_LOG_ENTRY_TYPE = "IMAGE"
const USER_LOG_ENTRY_TYPE = "USER"

func (this *App) LogUserEvent(r *http.Request, objType string, id int64, logType dao.ChangesLogEntryType, description string) {
	go func() {
		u := r.Context().Value(handler.USER_REQUEST_VARIABLE)
		if u != nil {
			u := u.(*dao.User)

			err := this.ChangesLogDao.Insert(dao.ChangesLogEntry{
				ObjectType:   objType,
				ObjectId:     id,
				AuthProvider: u.AuthProvider,
				ExtId:        u.ExtId,
				Login:        u.Info.Login,
				Type:         logType,
				Description:  description,
				Time:         dao.JSONTime(time.Now()),
			})
			if err != nil {
				log.Error("Can not add changelog entry!", err)
			}
		} else {
			log.Error("User is null but authorized!")
		}
	}()
}

func GetBoolParameter(req *http.Request, name string, _default bool) bool {
	valStr := req.FormValue(name)
	value, err := strconv.ParseBool(valStr)
	if err != nil {
		return _default
	}
	return value
}
