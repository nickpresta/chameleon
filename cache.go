package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
)

// CachedResponse respresents a response to be cached.
type CachedResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

// SpecResponse represents a specification for a response.
type SpecResponse struct {
	StatusCode  int               `json:"status_code"`
	ContentFile string            `json:"content"`
	Headers     map[string]string `json:"headers"`
}

// Spec represents a full specification to describe a response and how to look up its index.
type Spec struct {
	SpecResponse `json:"response"`
	Key          string `json:"key"`
}

// A Keyer interface is used to generate a key for a given request.
type Keyer interface {
	Key(r *http.Request) string
}

// A Cacher interface is used to provide a mechanism of storage for a given request and response.
type Cacher interface {
	Keyer
	Get(key string) *CachedResponse
	Put(key string, r *httptest.ResponseRecorder) *CachedResponse
}

type defaultKeyer struct {
}

func (k defaultKeyer) Key(r *http.Request) string {
	// TODO: Support custom hashers communicating via STDIN/STDOUT and called via os/exec.Output
	// This will have to do for now
	hasher := md5.New()
	key := r.URL.RequestURI() + r.Method
	hasher.Write([]byte(key))

	if r.Header.Get("chameleon-hash-body") != "" {
		var buf bytes.Buffer
		buf.ReadFrom(r.Body)
		bufBytes := buf.Bytes()

		_, err := io.Copy(hasher, bytes.NewReader(bufBytes))
		if err != nil {
			panic(err)
		}
		// Put the body back on the request so it can read again
		r.Body = ioutil.NopCloser(bytes.NewReader(bufBytes))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

type diskCacher struct {
	defaultKeyer
	cache    map[string]*CachedResponse
	dataDir  string
	specPath string
}

// NewDiskCacher creates a new disk cacher for a given data directory.
func NewDiskCacher(dataDir string) diskCacher {

	dc := diskCacher{
		cache:    nil,
		dataDir:  dataDir,
		specPath: path.Join(dataDir, "spec.json"),
	}

	cache := make(map[string]*CachedResponse)
	specs := dc.loadSpecs()

	for _, spec := range specs {
		body, err := ioutil.ReadFile(path.Join(dataDir, spec.SpecResponse.ContentFile))
		if err != nil {
			panic(err)
		}
		response := &CachedResponse{
			StatusCode: spec.StatusCode,
			Headers:    spec.Headers,
			Body:       body,
		}
		cache[spec.Key] = response
	}

	dc.cache = cache
	return dc
}

func (c diskCacher) Get(key string) *CachedResponse {
	return c.cache[key]
}

func (c diskCacher) loadSpecs() []Spec {
	specContent, err := ioutil.ReadFile(c.specPath)
	if err != nil {
		specContent = []byte{'[', ']'}
	}

	var specs []Spec
	err = json.Unmarshal(specContent, &specs)
	if err != nil {
		panic(err)
	}

	return specs
}

func (c diskCacher) Put(key string, resp *httptest.ResponseRecorder) *CachedResponse {
	specs := c.loadSpecs()

	specHeaders := make(map[string]string)
	for k, v := range resp.Header() {
		specHeaders[k] = strings.Join(v, ", ")
	}

	newSpec := Spec{
		Key: key,
		SpecResponse: SpecResponse{
			StatusCode:  resp.Code,
			ContentFile: key,
			Headers:     specHeaders,
		},
	}

	specs = append(specs, newSpec)

	contentFilePath := path.Join(c.dataDir, key)
	err := ioutil.WriteFile(contentFilePath, resp.Body.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	specBytes, err := json.MarshalIndent(specs, "", "    ")
	err = ioutil.WriteFile(c.specPath, specBytes, 0644)
	if err != nil {
		panic(err)
	}

	c.cache[key] = &CachedResponse{
		StatusCode: resp.Code,
		Headers:    specHeaders,
		Body:       resp.Body.Bytes(),
	}

	return c.cache[key]
}
