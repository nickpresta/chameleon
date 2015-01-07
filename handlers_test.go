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
	Headers:    map[string]string{"Foo": "Bar"},
}

type mockCacher struct {
}

func (m mockCacher) Get(key string) *CachedResponse {
	return nil
}

func (m mockCacher) Put(key string, r *httptest.ResponseRecorder) *CachedResponse {
	return fakeResp
}

func TestCachedProxyHandlerWithCmdHasher(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Foo", fakeResp.Headers["Foo"])
		w.WriteHeader(fakeResp.StatusCode)
		fmt.Fprintf(w, string(fakeResp.Body))
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	handler := CachedProxyHandler(
		ProxyHandler,
		serverURL,
		mockCacher{},
		DefaultHasher{},
	)

	w := httptest.NewRecorder()

	q := serverURL.Query()
	q.Set("q", "golang")
	serverURL.RawQuery = q.Encode()
	serverURL.Path = "/search"
	req, _ := http.NewRequest("POST", serverURL.String(), strings.NewReader("POST BODY"))
	req.Header.Set("Sample", "Header")
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
}
