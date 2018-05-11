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
)

type Pos struct {
	x string
	y string
	z string
	t string
}

type Mapping struct {
	pattern []*fasttemplate.Template
}

func (this *Mapping) mkUrl(z string, x string, y string) string {
	l := len(this.pattern)
	idx := rand.Intn(l)
	pattern := this.pattern[idx]
	return pattern.ExecuteString(map[string]interface{}{"x":x, "y":y, "z":z, })
}

type Handler struct {
	baseDir    string
	urlMapping map[string]Mapping
	client     *http.Client
	semaphore  chan string
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

	this.semaphore <- path
	defer func() {
		<-this.semaphore
	}()

	_, err := os.Stat(path)
	if err == nil {
		log.Debug(path + " was concurrently downloaded by somebody else")
		return
	}

	mapping, found := this.urlMapping[pos.t]
	if !found {
		OnError(w, errors.New(pos.t), "Can not find mapping for type", http.StatusBadRequest)
		return
	}
	url := mapping.mkUrl(pos.z, pos.x, pos.y)
	tmpFile, err := this.fetchUrlToTmpFile(w, url)
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

func main() {
	log.Infof("Starting wwmap")

	configuration := config.Load("").TileCache
	urlMapping := make(map[string]Mapping)
	for t, u := range configuration.Types {
		if len(u) == 0 {
			log.Warnf("No patterns for type %s", t)
			continue
		}
		p := make([]*fasttemplate.Template, len(u))
		for i, urlPatternStr := range u {
			p[i] = fasttemplate.New(urlPatternStr, "[[", "]]")
		}
		urlMapping[t] = Mapping{pattern:p}
	}
	handler := Handler{
		baseDir:configuration.BaseDir,
		urlMapping: urlMapping,
		client: &http.Client{},
		semaphore: make(chan string, 10),
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
