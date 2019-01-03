package handler

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/handler"
	"net/http"
	"fmt"
	"os"
	"github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
	"strconv"
	"github.com/and-hom/wwmap/lib/dao"
	"net/url"
	"github.com/and-hom/wwmap/backend/passport"
	"strings"
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
		ExtId:info.Id,
		AuthProvider:authProvider,
		Role:dao.USER,
		Info: dao.UserInfo{
			FirstName:info.FirstName,
			LastName:info.LastName,
			Login:info.Login,
		},
		SessionId: sessionId,
	})
}

func (this *App) ForRoles(payload handler.HandlerFunction, roles ...dao.Role) handler.HandlerFunction {
	if len(roles)==0 {
		return payload
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		if !this.CheckRoleAllowedAndMakeResponse(writer, request, roles...) {
			return
		}
		payload(writer, request)
	}
}

func (this *App) CheckRoleAllowedAndMakeResponse(w http.ResponseWriter, r *http.Request, allowedRoles ...dao.Role) bool {
	allowed, err := this.CheckRoleAllowed(r, allowedRoles...)
	if err != nil {
		OnError500(w, err, "Can not check permissions")
		return false
	}
	if !allowed {
		msg := ""
		if len(allowedRoles) == 1 {
			msg = fmt.Sprintf("Sorry! You haven't role %s", allowedRoles[0])
		} else {
			msg = fmt.Sprintf("Sorry! You haven't any of following roles: %s", dao.Join(", ", allowedRoles...))
		}
		OnError(w, nil, msg, http.StatusUnauthorized)
		return false
	}
	return true
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

func (this *App) CheckRoleAllowed(r *http.Request, allowedRoles ...dao.Role) (bool, error) {
	authProvider, info, err := this.GetUserInfo(r)
	if err != nil {
		return false, err
	}

	role, err := this.UserDao.GetRole(authProvider, info.Id)
	if err != nil {
		return false, err
	}
	for i := 0; i < len(allowedRoles); i++ {
		if allowedRoles[i] == role {
			return true, nil
		}
	}
	return false, nil
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