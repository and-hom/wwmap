package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	. "github.com/and-hom/wwmap/lib/handler"
	. "github.com/and-hom/wwmap/lib/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
	"time"
)

var GMT_LOC, _ = time.LoadLocation("UTC")
var version string = "development"

type Pos struct {
	x int
	y int
	z int
	t string
}

func (this Pos) String() string {
	return fmt.Sprintf("%s/%d/%d/%d", this.t, this.z, this.x, this.y)
}

type CdnUrlPattern struct {
	template  *template.Template
	semaphore chan int
}

func (this CdnUrlPattern) mkUrl(pos Pos) string {
	urlStrBuf := bytes.Buffer{}
	err := this.template.Execute(&urlStrBuf, map[string]interface{}{
		"x": pos.x,
		"y": pos.y,
		"z": pos.z,
	})
	if err != nil {
		log.Error("Can not create url", pos, err)
	}
	return urlStrBuf.String()
}

type Mapping struct {
	cdn []CdnUrlPattern
}

func (this *Mapping) mkUrl(pos Pos) LockableUrlHolder {
	l := len(this.cdn)
	idx := rand.Intn(l)
	cdn := this.cdn[idx]

	cdn.semaphore <- 0
	return LockableUrlHolder{
		Url: cdn.mkUrl(pos),
		Unlock: func() {
			<-cdn.semaphore
		},
	}
}

type LockableUrlHolder struct {
	Url    string
	Unlock func()
}

type PathLock struct {
	key        string
	mutexByUrl *sync.Map
	mutex      *sync.Mutex
}

func (this *PathLock) Lock() {
	foundUrlMutex, _ := this.mutexByUrl.LoadOrStore(this.key, &sync.Mutex{})
	this.mutex = foundUrlMutex.(*sync.Mutex)
	this.mutex.Lock()
}

func (this *PathLock) Unlock() {
	this.mutexByUrl.Delete(this.key)
	this.mutex.Unlock()
}

type DataHandler struct {
	Handler
	baseDir     string
	urlMapping  map[string]Mapping
	client      *http.Client
	mutexByUrl  sync.Map
	defaultData []byte
	version     string
}

func (this *DataHandler) Init() {
	this.Register("/{type}/{z}/{x}/{y}.png", HandlerFunctions{Get: this.tile, Head: this.tile})
	this.Register("/version", HandlerFunctions{Get: this.Version})
}

func (this *DataHandler) initDefaultImageIfNecessary() {
	if this.defaultData == nil {
		log.Info("Generate empty png for missing tiles")
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.Set(0, 0, color.RGBA{0, 0, 0, 0})
		buf := bytes.Buffer{}
		png.Encode(&buf, img)
		this.defaultData = buf.Bytes()
	}
}

func (this *DataHandler) fetchUrlToTmpFile(w http.ResponseWriter, url string) (string, error) {
	log.Info("Fetch ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		OnError500(w, err, "Can not create request for "+url)
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/601.7.7 (KHTML, like Gecko) Version/9.1.2 Safari/601.7.7")

	resp, err := this.client.Do(req)
	if err != nil {
		OnError500(w, err, "Can not fetch tile file "+url)
		return "", err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		err := errors.Errorf("GET %s returned %d", url, resp.StatusCode)
		OnError(w, err, "Can not fetch tile file "+url, resp.StatusCode)
		return "", err
	}
	defer resp.Body.Close()

	f, err := ioutil.TempFile(os.TempDir(), "wwmap")
	if err != nil {
		OnError500(w, err, "Can not create tile file ")
		return "", err
	}
	defer f.Close()

	if resp.StatusCode == http.StatusOK {
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			OnError500(w, err, "Can not write tile file "+f.Name())
			return "", err
		}
	} else {
		this.initDefaultImageIfNecessary()
		log.Info("Use empty image for missing tile")
		w.Write(this.defaultData)
	}

	return f.Name(), nil
}

func (this *DataHandler) fetch(w http.ResponseWriter, pos Pos) error {
	path := this.cachePath(pos)

	pathLock := PathLock{key: path, mutexByUrl: &this.mutexByUrl}
	pathLock.Lock()
	defer pathLock.Unlock()

	mapping, found := this.urlMapping[pos.t]
	url := mapping.mkUrl(pos)
	defer url.Unlock()

	_, err := os.Stat(path)
	if err == nil {
		log.Info(path + " was concurrently downloaded by somebody else")
		return err
	}

	time.Sleep(time.Second)

	if !found {
		err := errors.New(pos.t)
		OnError(w, err, "Can not find mapping for type", http.StatusBadRequest)
		return err
	}
	tmpFile, err := this.fetchUrlToTmpFile(w, url.Url)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir|0777)
	if err != nil {
		OnError500(w, err, "Can not create parent directory for file "+path)
		return err
	}
	err = os.Rename(tmpFile, path)
	if err != nil {
		OnError500(w, err, "Can not create file "+path)
		return err
	}
	defer os.Remove(tmpFile)
	return nil
}

func (this *DataHandler) tile(w http.ResponseWriter, req *http.Request) {
	pathParams := mux.Vars(req)

	x, err := strconv.ParseInt(pathParams["x"], 10, 32)
	if err != nil {
		OnError(w, err, "x should be an integer", http.StatusBadRequest)
	}
	y, err := strconv.ParseInt(pathParams["y"], 10, 32)
	if err != nil {
		OnError(w, err, "y should be an integer", http.StatusBadRequest)
	}
	z, err := strconv.ParseInt(pathParams["z"], 10, 32)
	if err != nil {
		OnError(w, err, "z should be an integer", http.StatusBadRequest)
	}
	pos := Pos{
		t: pathParams["type"],
		x: int(x),
		y: int(y),
		z: int(z),
	}

	path := this.cachePath(pos)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		err := this.fetch(w, pos)
		if err != nil {
			return
		}
	} else if err != nil {
		OnError500(w, err, "Can not get stat for tile file "+path)
		return
	}

	stat, err := os.Stat(path)
	if err != nil {
		OnError500(w, err, "Can not get stat for tile file "+path)
		return
	}
	modTime := stat.ModTime().In(GMT_LOC)
	w.Header().Add("Last-Modified", modTime.Format(http.TimeFormat))
	w.Header().Add("Expires", modTime.Add(24*time.Hour).Format(http.TimeFormat))
	w.Header().Add("Cache-Control", "public")

	ifModSinceStr := req.Header.Get("If-Modified-Since")
	ifModSince, err := time.Parse(http.TimeFormat, ifModSinceStr)
	if err != nil {
		ifModSince = time.Unix(0, 0)
	}

	if stat.ModTime().Before(ifModSince) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if req.Method == "GET" {
		w.Header().Add("Content-Length", fmt.Sprintf("%d", stat.Size()))
		f, err := os.Open(path)
		if err != nil {
			OnError500(w, err, "Can not read tile file "+path)
			return
		}
		defer f.Close()
		io.Copy(w, f)
	} else {
		w.Header().Add("Content-Length", "0")
	}
}

func (this *DataHandler) cachePath(pos Pos) string {
	return fmt.Sprintf("%s/%s/%d/%d/%d.png", this.baseDir, pos.t, pos.z, pos.x, pos.y)
}

func typeCdnMapping(configuration config.TileCache) map[string]Mapping {
	typeCdnMapping := make(map[string]Mapping)
	funcMap := template.FuncMap{
		"div": func(x int, y int) int {
			return x / y
		},
		"sum": func(x int, y int) int {
			return x + y
		},
	}
	for t, u := range configuration.Types {
		if len(u) == 0 {
			log.Warnf("No patterns for type %s", t)
			continue
		}

		p := make([]CdnUrlPattern, len(u))
		for i, urlPatternStr := range u {
			tmpl, err := template.New(fmt.Sprintf("%s-%d", t, i)).Funcs(funcMap).Parse(urlPatternStr)
			if err != nil {
				log.Fatalf("Can not process template %s %v+", urlPatternStr, err)
			}
			p[i] = CdnUrlPattern{
				template:  tmpl,
				semaphore: make(chan int, 5),
			}
		}

		typeCdnMapping[t] = Mapping{cdn: p}
	}
	return typeCdnMapping
}

func main() {
	log.Infof("Starting wwmap")

	fullConfiguration := config.Load("")
	fullConfiguration.ChangeLogLevel()
	configuration := fullConfiguration.TileCache

	r := mux.NewRouter()

	handler := DataHandler{
		Handler:    Handler{R: r},
		baseDir:    configuration.BaseDir,
		urlMapping: typeCdnMapping(configuration),
		client: &http.Client{
			Timeout: 4 * time.Second,
		},
		version: version,
	}

	handler.Init()

	log.Infof("Starting tiles server on %v+", configuration.BindTo)

	srv := &http.Server{
		ReadTimeout: 5 * time.Second,
		Addr:        configuration.BindTo,
		Handler:     WrapWithLogging(r, fullConfiguration),
	}
	if configuration.ReadTimeout > 0 {
		srv.ReadTimeout = configuration.ReadTimeout
	}
	if configuration.WriteTimeout > 0 {
		srv.WriteTimeout = configuration.WriteTimeout
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}

func (this *DataHandler) Version(w http.ResponseWriter, req *http.Request) {
	this.JsonAnswer(w, this.version)
}
