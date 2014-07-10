package cache

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
)

type CachedResponse struct {
}

type Keyer interface {
	Key(r *http.Request) string
}

type Cacher interface {
	Keyer
	Get(key string) *http.Response
	Put(key string, r *httptest.ResponseRecorder)
}

type DefaultKeyer struct {
}

func (k DefaultKeyer) Key(r *http.Request) string {
	key := r.RequestURI + r.Method
	if strings.ToLower(r.Header.Get("chameleon-hash-body")) == "true" {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		key += string(body)
	}
	return key
}

type diskCacher struct {
	DefaultKeyer
	cache    map[string][]byte
	dataDir  string
	specPath string
}

func NewDiskCacher(dataDir string) diskCacher {
	return diskCacher{
		cache:    make(map[string][]byte),
		dataDir:  dataDir,
		specPath: path.Join(dataDir, "spec.json"),
	}
}

func (c diskCacher) Get(key string) *http.Response {
	return nil
}

func (c diskCacher) Put(key string, resp *httptest.ResponseRecorder) {
	contentFilePath := path.Join(c.dataDir, key)
	err := ioutil.WriteFile(contentFilePath, resp.Body.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}
