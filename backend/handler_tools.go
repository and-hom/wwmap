package main

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"fmt"
	"io"
	"os"
	"errors"
	. "github.com/and-hom/wwmap/lib/geoparser"
	"encoding/json"
	"github.com/and-hom/wwmap/lib/geo"
	. "github.com/and-hom/wwmap/lib/http"
	"strconv"
	"github.com/and-hom/wwmap/lib/dao"
	"net/url"
)

func (this *Handler) JsonpAnswer(callback string, object interface{}, _default string) []byte {
	return []byte(callback + "(" + this.JsonStr(object, _default) + ");")
}

func (this *Handler) JsonStr(f interface{}, _default string) string {
	bytes, err := json.Marshal(f)
	if err != nil {
		log.Errorf("Can not serialize object %v: %s", f, err.Error())
		return _default
	}
	return string(bytes)
}


func (this *Handler) JsonAnswer(w http.ResponseWriter, f interface{}) {
	bytes, err := json.Marshal(f)
	if err != nil {
		OnError500(w, err, fmt.Sprintf("Can not serialize object %v", f))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(bytes)
}

func (this *Handler) CorsGetOptionsStub(w http.ResponseWriter, r *http.Request) {
	CorsHeaders(w, POST, GET, OPTIONS, PUT, DELETE)
	// for cors only
}

func (this *Handler) bboxFormValue(w http.ResponseWriter, req *http.Request) (geo.Bbox, error) {
	bboxStr := req.FormValue("bbox")
	bbox, err := geo.NewBbox(bboxStr)
	if err != nil {
		OnError(w, err, fmt.Sprintf("Can not parse bbox: %v", bbox), http.StatusBadRequest)
		return geo.Bbox{}, err
	}
	return bbox, nil
}

func (this *Handler) tileParams(w http.ResponseWriter, req *http.Request) (string, geo.Bbox, error) {
	callback := req.FormValue("callback")
	bbox, err := this.bboxFormValue(w, req)
	if err != nil {
		return "", geo.Bbox{}, err
	}

	return callback, bbox, nil
}

func (this *Handler) tileParamsZ(w http.ResponseWriter, req *http.Request) (string, geo.Bbox, int, error) {

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

func (this *Handler) CloseAndRemove(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}

func (this *Handler) geoParser(r io.ReadSeeker) (GeoParser, error) {
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

func (this *Handler) CreateMissingUser(r *http.Request) error {
	token := GetOauthToken(r)
	info, err := this.yandexPassport.ResolveUserInfo(token)
	if err != nil {
		return err
	}

	return this.userDao.CreateIfNotExists(dao.User{
		YandexId:info.Id,
		Role:dao.USER,
		Info: dao.UserInfo{
			FirstName:info.FirstName,
			LastName:info.LastName,
			Login:info.Login,
		},

	})
}

func (this *Handler) CheckRoleAllowedAndMakeResponse(w http.ResponseWriter, r *http.Request, allowedRoles ...dao.Role) bool {
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

func (this *Handler) CheckRoleAllowed(r *http.Request, allowedRoles ...dao.Role) (bool, error) {
	token := GetOauthToken(r)
	info, err := this.yandexPassport.ResolveUserInfo(token)
	if err != nil {
		return false, err
	}

	role, err := this.userDao.GetRole(info.Id)
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

func (this *Handler) collectReferer(r *http.Request) {
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

	this.refererStorage.Put(refererUrl)
}