package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"github.com/and-hom/wwmap/lib/config"
	. "github.com/and-hom/wwmap/lib/http"
	"io"
	"math/rand"
	"github.com/valyala/fasttemplate"
	"github.com/pkg/errors"
	"path/filepath"
	"io/ioutil"
	"time"
	"sync"
)

type Pos struct {
	x string
	y string
	z string
	t string
}

type CdnUrlPattern struct {
	template  *fasttemplate.Template
	semaphore chan int
}

func (this CdnUrlPattern) mkUrl(pos Pos) string {
	return this.template.ExecuteString(map[string]interface{}{"x":pos.x, "y":pos.y, "z":pos.z, })
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

type Handler struct {
	baseDir    string
	urlMapping map[string]Mapping
	client     *http.Client
	mutexByUrl sync.Map
}

func (this *Handler) fetchUrlToTmpFile(w http.ResponseWriter, url string) (string, error) {
	log.Info("Fetch ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		OnError500(w, err, "Can not create request for " + url)
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/601.7.7 (KHTML, like Gecko) Version/9.1.2 Safari/601.7.7")

	resp, err := this.client.Do(req)
	if err != nil {
		OnError500(w, err, "Can not fetch tile file " + url)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		err := errors.Errorf("GET %s returned %d", url, resp.StatusCode)
		OnError500(w, err, "Can not fetch tile file " + url)
		return "", err
	}
	defer resp.Body.Close()

	f, err := ioutil.TempFile(os.TempDir(), "wwmap")
	if err != nil {
		OnError500(w, err, "Can not create tile file ")
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		OnError500(w, err, "Can not write tile file " + f.Name())
		return "", err
	}

	return f.Name(), nil
}

func (this *Handler) fetch(w http.ResponseWriter, pos Pos) {
	path := this.cachePath(pos)

	pathLock := PathLock{key:path, mutexByUrl:&this.mutexByUrl}
	pathLock.Lock()
	defer pathLock.Unlock()

	mapping, found := this.urlMapping[pos.t]
	url := mapping.mkUrl(pos)
	defer url.Unlock()

	_, err := os.Stat(path)
	if err == nil {
		log.Info(path + " was concurrently downloaded by somebody else")
		return
	}

	time.Sleep(time.Second)

	if !found {
		OnError(w, errors.New(pos.t), "Can not find mapping for type", http.StatusBadRequest)
		return
	}
	tmpFile, err := this.fetchUrlToTmpFile(w, url.Url)
	if err != nil {
		return
	}
	err = os.MkdirAll(filepath.Dir(path), os.ModeDir | 0777)
	if err != nil {
		OnError500(w, err, "Can not create parent directory for file " + path)
		return
	}
	err = os.Rename(tmpFile, path)
	if err != nil {
		OnError500(w, err, "Can not create file " + path)
		return
	}
	defer os.Remove(tmpFile)
}

func (this *Handler) tile(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, "GET, HEAD")

	pathParams := mux.Vars(req)

	pos := Pos{
		t : pathParams["type"],
		x : pathParams["x"],
		y : pathParams["y"],
		z : pathParams["z"],
	}

	path := this.cachePath(pos)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		this.fetch(w, pos)
	} else if err != nil {
		OnError500(w, err, "Can not get stat for tile file " + path)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		OnError500(w, err, "Can not read tile file " + path)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func (this *Handler) cachePath(pos Pos) string {
	return this.baseDir + "/" + pos.t + "/" + pos.z + "/" + pos.x + "/" + pos.y + ".png"
}

func typeCdnMapping(configuration config.TileCache) map[string]Mapping {
	typeCdnMapping := make(map[string]Mapping)
	for t, u := range configuration.Types {
		if len(u) == 0 {
			log.Warnf("No patterns for type %s", t)
			continue
		}

		p := make([]CdnUrlPattern, len(u))
		for i, urlPatternStr := range u {
			p[i] = CdnUrlPattern{
				template: fasttemplate.New(urlPatternStr, "[[", "]]"),
				semaphore:make(chan int, 3),
			}
		}

		typeCdnMapping[t] = Mapping{cdn:p}
	}
	return typeCdnMapping
}

func main() {
	log.Infof("Starting wwmap")

	configuration := config.Load("").TileCache

	handler := Handler{
		baseDir:configuration.BaseDir,
		urlMapping: typeCdnMapping(configuration),
		client: &http.Client{
			Timeout: 4 * time.Second,
		},
	}

	r := mux.NewRouter()
	r.HandleFunc("/{type}/{z}/{x}/{y}.png", handler.tile)

	log.Infof("Starting tiles server on %v+", configuration.BindTo)
	http.Handle("/", r)
	err := http.ListenAndServe(configuration.BindTo, http.DefaultServeMux)
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
}
