package handler

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"fmt"
	"io"
	"os"
	"errors"
	. "github.com/and-hom/wwmap/lib/geoparser"
	"github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
	"strconv"
	"github.com/and-hom/wwmap/lib/dao"
	"net/url"
)

func (this *App) bboxFormValue(w http.ResponseWriter, req *http.Request) (geo.Bbox, error) {
	bboxStr := req.FormValue("bbox")
	bbox, err := geo.NewBbox(bboxStr)
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

func (this *App) geoParser(r io.ReadSeeker) (GeoParser, error) {
	gpxParser, err := InitGpxParser(r)
	if err == nil {
		return gpxParser, nil
	}
	log.Warn(err)
	r.Seek(0, 0)
	kmlParser, err := InitKmlParser(r)
	if err == nil {
		return kmlParser, nil
	}
	log.Warn(err)
	return nil, errors.New("Can not find valid parser for this format!")
}

func (this *App) CreateMissingUser(r *http.Request) error {
	token := GetOauthToken(r)
	info, err := this.YandexPassport.ResolveUserInfo(token)
	if err != nil {
		return err
	}

	return this.UserDao.CreateIfNotExists(dao.User{
		YandexId:info.Id,
		Role:dao.USER,
		Info: dao.UserInfo{
			FirstName:info.FirstName,
			LastName:info.LastName,
			Login:info.Login,
		},

	})
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
			fmt.Sprintf("Sorry! You haven't role %s", allowedRoles[0])
		} else {
			fmt.Sprintf("Sorry! You haven't any of following roles: %s", dao.Join(", ", allowedRoles...))
		}
		OnError(w, nil, msg, http.StatusUnauthorized)
		return false
	}
	return true
}

func (this *App) CheckRoleAllowed(r *http.Request, allowedRoles ...dao.Role) (bool, error) {
	token := GetOauthToken(r)
	info, err := this.YandexPassport.ResolveUserInfo(token)
	if err != nil {
		return false, err
	}

	role, err := this.UserDao.GetRole(info.Id)
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
	if refererUrl.Hostname() == "localhost" {
		return
	}

	err = this.RefererStorage.Put(refererUrl)
	if err != nil {
		log.Error("Can not store referer ", err)
	}
}