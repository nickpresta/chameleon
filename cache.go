package main

import (
	"encoding/json"
	"io/ioutil"
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

// A Cacher interface is used to provide a mechanism of storage for a given request and response.
type Cacher interface {
	Get(key string) *CachedResponse
	Put(key string, r *httptest.ResponseRecorder) *CachedResponse
}

// DiskCacher is the default cacher which writes to disk
type DiskCacher struct {
	cache    map[string]*CachedResponse
	dataDir  string
	specPath string
}

// NewDiskCacher creates a new disk cacher for a given data directory.
func NewDiskCacher(dataDir string) DiskCacher {

	dc := DiskCacher{
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

// Get fetches a CachedResponse for a given key
func (c DiskCacher) Get(key string) *CachedResponse {
	return c.cache[key]
}

func (c DiskCacher) loadSpecs() []Spec {
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

// Put stores a CachedResponse for a given key and response
func (c DiskCacher) Put(key string, resp *httptest.ResponseRecorder) *CachedResponse {
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
