package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockFileSystem struct {
}

func (fs mockFileSystem) WriteFile(path string, content []byte) error {
	return nil
}

func (fs mockFileSystem) ReadFile(path string) ([]byte, error) {
	if strings.HasSuffix(path, "-error") {
		return nil, fmt.Errorf("SOMETHING BROKE")
	}

	// Return specs when spec.json is requested (which should be always)
	if strings.HasSuffix(path, "spec.json") {
		specs := []Spec{
			Spec{
				Key: "key",
				SpecResponse: SpecResponse{
					StatusCode:  418,
					ContentFile: "key",
					Headers:     map[string]string{"Content-Type": "text/plain"},
				},
			},
		}
		var specsContent bytes.Buffer
		dec := json.NewEncoder(&specsContent)
		_ = dec.Encode(specs)
		return specsContent.Bytes(), nil
	}
	// Otherwise we're returning a file that has been cached
	return []byte("CACHED CONTENT FILE"), nil
}

func TestDiskCacherGet(t *testing.T) {
	cacher := NewDiskCacher("")
	cacher.FileSystem = mockFileSystem{}
	cacher.SeedCache()

	response := cacher.Get("key")
	if response.StatusCode != 418 {
		t.Errorf("Got: `%v`; Expected: `418`", response.StatusCode)
	}
	if response.Headers["Content-Type"] != "text/plain" {
		t.Errorf("Got: `%v`; Expected: `text/plain`", response.Headers["Content-Type"])
	}
	if string(response.Body) != "CACHED CONTENT FILE" {
		t.Errorf("Got: `%v`; Expected: `CACHED CONTENT FILE`", string(response.Body))
	}
}

func TestDiskCacherPut(t *testing.T) {
	cacher := NewDiskCacher("")
	cacher.FileSystem = mockFileSystem{}
	cacher.SeedCache()

	recorder := httptest.NewRecorder()
	var body bytes.Buffer
	_, _ = body.WriteString("THIS IS A NEW BODY")
	recorder.Header().Set("Content-Type", "text/plain")
	recorder.Code = 700
	recorder.Body = &body
	response := cacher.Put("new_key", recorder)

	if response.StatusCode != 700 {
		t.Errorf("Got: `%v`; Expected: `700`", response.StatusCode)
	}
	if response.Headers["Content-Type"] != "text/plain" {
		t.Errorf("Got: `%v`; Expected: `text/plain`", response.Headers["Content-Type"])
	}
	if string(response.Body) != "THIS IS A NEW BODY" {
		t.Errorf("Got: `%v`; Expected: `THIS IS A NEW BODY`", string(response.Body))
	}
}

func TestDiskCacherSeedCacheNoSpecs(t *testing.T) {
	cacher := NewDiskCacher("")
	cacher.FileSystem = mockFileSystem{}
	cacher.specPath = "-error"

	cacher.SeedCache()
	if len(cacher.cache) != 0 {
		t.Errorf("Got: `%v`; Expected: `0`", len(cacher.cache))
	}
}

func TestDiskCacherPutSkipDiskSeeded(t *testing.T) {
	cacher := NewDiskCacher("")
	cacher.FileSystem = mockFileSystem{}
	cacher.SeedCache()

	recorder := httptest.NewRecorder()
	recorder.Header().Set("_chameleon-seeded-skip-disk", "true")
	response := cacher.Put("new_key", recorder)

	if _, ok := response.Headers["_chameleon-seeded-skip-disk"]; ok {
		t.Errorf("Unexpected header `_chameleon-seeded-skip-disk`")
	}
}
