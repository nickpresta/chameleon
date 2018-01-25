package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

var fakeResp = &CachedResponse{
	StatusCode: 418,
	Body:       []byte("Hello, World!"),
	Headers:    map[string]string{"Foo": "Bar", "chameleon-request-hash": "abcdef12345"},
}

type mockCacher struct {
	data map[string]*CachedResponse
}

func (m mockCacher) Get(key string) *CachedResponse {
	return m.data[key]
}

func (m mockCacher) Put(key string, r *httptest.ResponseRecorder) *CachedResponse {
	specHeaders := make(map[string]string)
	for k, v := range r.Header() {
		specHeaders[k] = strings.Join(v, ", ")
	}

	m.data[key] = &CachedResponse{
		StatusCode: r.Code,
		Body:       r.Body.Bytes(),
		Headers:    specHeaders,
	}
	return m.data[key]
}

func TestCachedProxyHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Foo", fakeResp.Headers["Foo"])
		w.WriteHeader(fakeResp.StatusCode)
		fmt.Fprintf(w, string(fakeResp.Body))
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	handler := CachedProxyHandler(
		serverURL,
		mockCacher{data: make(map[string]*CachedResponse)},
		DefaultHasher{},
		ProxyHandler(false),
	)

	w := httptest.NewRecorder()

	q := serverURL.Query()
	q.Set("q", "golang")
	serverURL.RawQuery = q.Encode()
	serverURL.Path = "/search"
	req, _ := http.NewRequest("POST", serverURL.String(), strings.NewReader("POST BODY"))
	req.Header.Set("Sample", "Header")
	req.Header.Set("chameleon-request-hash", fakeResp.Headers["chameleon-request-hash"])
	handler.ServeHTTP(w, req)

	// Check that the Proxy worked (response is the same as request)
	if w.Code != fakeResp.StatusCode {
		t.Errorf("Got: `%v`; Expected: `%v`", w.Code, fakeResp.StatusCode)
	}
	if w.Header().Get("Foo") != fakeResp.Headers["Foo"] {
		t.Errorf("Got: `%v`; Expected: `%v`", w.Header().Get("Foo"), fakeResp.Headers["Foo"])
	}
	body, _ := ioutil.ReadAll(w.Body)
	if !bytes.Equal(body, fakeResp.Body) {
		t.Errorf("Got: `%v`; Expected: `%v`", string(body), string(fakeResp.Body))
	}
	if w.Header().Get("chameleon-request-hash") == "" {
		t.Errorf("Hash was not returned with response.")
	}
	if w.Header().Get("chameleon-request-hash") != fakeResp.Headers["chameleon-request-hash"] {
		t.Errorf("Got: `%v`; Expected: `%v`", w.Header().Get("chameleon-request-hash"), fakeResp.Headers["chameleon-request-hash"])
	}
}

func TestPreseedHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not have hit the server. Response was preseeded")
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	cache := mockCacher{data: make(map[string]*CachedResponse)}
	cachedProxyHandler := CachedProxyHandler(
		serverURL,
		cache,
		DefaultHasher{},
		ProxyHandler(false),
	)
	preseedHandler := PreseedHandler(
		cache,
		DefaultHasher{},
	)

	// Seed /foobar
	serverURL.Path = "/_seed"
	req, _ := http.NewRequest("POST", serverURL.String(), strings.NewReader(
		`{
			"Request": {
				"URL": "/foobar",
				"Method": "GET",
				"Body": ""
			},
			"Response": {
				"Body": "FOOBAR BODY",
				"StatusCode": 942,
				"Headers": {
					"Content-Type": "application/json"
				}
			}
		}`,
	))
	w := httptest.NewRecorder()
	preseedHandler.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Errorf("Got: `%v`; Expected: `201`; Error was `%v`", w.Code, w.Body.String())
	}

	serverURL.Path = "/foobar"
	req, _ = http.NewRequest("GET", serverURL.String(), nil)
	w = httptest.NewRecorder()
	cachedProxyHandler.ServeHTTP(w, req)

	if w.Body.String() != "FOOBAR BODY" {
		t.Errorf("Got: `%v`; Expected: `FOOBAR BODY`", w.Body.String())
	}

	if w.Code != 942 {
		t.Errorf("Got: `%v`; Expected: `942`", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Got: `%v`; Expected: `application/json`", w.Header().Get("Content-Type"))
	}
	if w.Header().Get("chameleon-request-hash") == "" {
		t.Errorf("Hash was not returned with response.")
	}
}

func TestPreseedHandlerWithRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		if string(body) != `{"post":"body"}` {
			t.Errorf("Got: `%v`; Expected `{\"post\":\"body\"}`", string(body))
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	cache := mockCacher{data: make(map[string]*CachedResponse)}
	cachedProxyHandler := CachedProxyHandler(
		serverURL,
		cache,
		DefaultHasher{},
		ProxyHandler(false),
	)
	preseedHandler := PreseedHandler(
		cache,
		DefaultHasher{},
	)

	// Seed /foobar
	serverURL.Path = "/_seed"
	req, _ := http.NewRequest("POST", serverURL.String(), strings.NewReader(
		`{
			"Request": {
				"URL": "/foobar",
				"Method": "POST",
				"Body": "{\"foo\":\"bar\"}"
			},
			"Response": {
				"Body": "FOOBAR BODY",
				"StatusCode": 942,
				"Headers": {
					"Content-Type": "application/json"
				}
			}
		}`,
	))
	w := httptest.NewRecorder()
	preseedHandler.ServeHTTP(w, req)

	serverURL.Path = "/foobar"
	req, _ = http.NewRequest("POST", serverURL.String(), strings.NewReader(`{"foo":"bar"}`))
	w = httptest.NewRecorder()
	cachedProxyHandler.ServeHTTP(w, req)

	req, _ = http.NewRequest("POST", serverURL.String(), strings.NewReader(`{"post":"body"}`))
	w = httptest.NewRecorder()
	cachedProxyHandler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Server wasn't hit with the correct body")
	}
	if w.Header().Get("chameleon-request-hash") == "" {
		t.Errorf("Hash was not returned with response.")
	}
}

func TestPreseedHandlerBadJSON(t *testing.T) {
	preseedHandler := PreseedHandler(
		mockCacher{},
		DefaultHasher{},
	)

	req, _ := http.NewRequest("POST", "/_seed", strings.NewReader("BAD JSON"))
	w := httptest.NewRecorder()
	preseedHandler.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Got: `%v`; Expected: `500`", w.Code)
	}
	if w.Header().Get("chameleon-request-hash") != "" {
		t.Errorf("Hash was returned for bad json.")
	}
}

func TestPreseedHandlerCachesDuplicateRequest(t *testing.T) {
	preseedHandler := PreseedHandler(
		mockCacher{data: make(map[string]*CachedResponse)},
		DefaultHasher{},
	)

	payload := `{
		"Request": {
			"URL": "/foobar",
			"Method": "GET",
			"Body": ""
		},
		"Response": {
			"Body": "FOOBAR BODY",
			"StatusCode": 942,
			"Headers": {
				"Content-Type": "application/json"
			}
		}
	}`

	req, _ := http.NewRequest("POST", "/_seed", strings.NewReader(payload))
	w := httptest.NewRecorder()
	preseedHandler.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Errorf("Got: `%v`; Expected: `201`", w.Code)
	}

	req, _ = http.NewRequest("POST", "/_seed", strings.NewReader(payload))
	w = httptest.NewRecorder()
	preseedHandler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Got: `%v`; Expected: `200`", w.Code)
	}
	if w.Header().Get("chameleon-request-hash") == "" {
		t.Errorf("Hash was not returned with response.")
	}
}

func TestPreseedHandlerBadURL(t *testing.T) {
	preseedHandler := PreseedHandler(
		mockCacher{},
		DefaultHasher{},
	)

	payload := `{
		"Request": {
			"URL": "%&%",
			"Method": "GET",
			"Body": ""
		},
		"Response": {
			"Body": "FOOBAR BODY",
			"StatusCode": 942,
			"Headers": {
				"Content-Type": "application/json"
			}
		}
	}`

	req, _ := http.NewRequest("POST", "/_seed", strings.NewReader(payload))
	w := httptest.NewRecorder()
	preseedHandler.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Got: `%v`; Expected: `500`", w.Code)
	}
	if w.Header().Get("chameleon-request-hash") != "" {
		t.Errorf("Hash was returned for bad url.")
	}
}
