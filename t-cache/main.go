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
)

type Pos struct {
	x string
	y string
	z string
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
}

func (this *Handler) fetch(w http.ResponseWriter, path string, url string) error {
	log.Info("Fetch ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		OnError500(w, err, "Can not create request for " + url)
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/601.7.7 (KHTML, like Gecko) Version/9.1.2 Safari/601.7.7")

	resp, err := this.client.Do(req)
	if err != nil {
		OnError500(w, err, "Can not fetch tile file " + path)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err := errors.Errorf("GET %s returned %d", url, resp.StatusCode)
		OnError500(w, err, "Can not fetch tile file " + path)
		return err
	}
	defer resp.Body.Close()

	err = os.MkdirAll(filepath.Dir(path), os.ModeDir | 0777)
	if err != nil {
		OnError500(w, err, "Can not create parent directory for file " + path)
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		OnError500(w, err, "Can not create tile file " + path)
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		OnError500(w, err, "Can not write tile file " + path)
		return err
	}
	return nil
}

func (this *Handler) tile(w http.ResponseWriter, req *http.Request) {
	CorsHeaders(w, "GET, HEAD")

	pathParams := mux.Vars(req)

	t := pathParams["type"]
	x := pathParams["x"]
	y := pathParams["y"]
	z := pathParams["z"]

	path := this.cachePath(t, z, x, y)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		mapping, found := this.urlMapping[t]
		if !found {
			OnError(w, errors.New(t), "Can not find mapping for type", http.StatusBadRequest)
			return
		}
		url := mapping.mkUrl(z, x, y)
		if this.fetch(w, path, url) != nil {
			return
		}
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

func (this *Handler) cachePath(t string, z string, x string, y string) string {
	return this.baseDir + "/" + t + "/" + z + "/" + x + "/" + y + ".png"
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
