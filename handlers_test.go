package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"testing"
)

func init() {
	//log.SetOutput(ioutil.Discard)
}

var fakeResp = &CachedResponse{
	StatusCode: 418,
	Body:       []byte("Hello, World!"),
	Headers:    map[string]string{"Foo": "Bar"},
}

type mockCacher struct{}

func (m mockCacher) Get(key string) *CachedResponse {
	return nil
}

func (m mockCacher) Put(key string, r *httptest.ResponseRecorder) *CachedResponse {
	return fakeResp
}

type testCommander struct {
	DefaultCommander
	stdin *bytes.Buffer
}

func (c testCommander) NewCmd(command string, stderr io.Writer, stdin io.Reader) *exec.Cmd {
	cmd := c.DefaultCommander.NewCmd(command, stderr, stdin)
	// Copy the STDIN sent to "command" to our bytes.Buffer for inspection later
	cmd.Stdin = io.TeeReader(cmd.Stdin, c.stdin)
	return cmd
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
	var stdin bytes.Buffer
	handler := CachedProxyHandler(
		ProxyHandler,
		serverURL,
		mockCacher{},
		CmdHasher{Command: "/bin/cat", Commander: testCommander{stdin: &stdin}},
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Sample", "Header")
	handler.ServeHTTP(w, req)

	d := json.NewDecoder(&stdin)
	var serializedReq serializedRequest
	err := d.Decode(&serializedReq)
	if err != nil { // Tests that STDIN is valid JSON
		t.Fatalf("Decoding of serialized request: %v", err)
	}

	log.Printf("REQ: %+v\n", serializedReq)

	// Check all fields sent in STDIN payload
	if serializedReq.Method != req.Method {
		t.Errorf("Method = '%v', want %v", serializedReq.Method, req.Method)
	}
	if serializedReq.Headers.Get("Sample") != "Header" {
		t.Errorf("Sample header = '%v', want Header", serializedReq.Headers.Get("Sample"))
	}

	// Check that the Proxy worked (response is the same as request)
	if w.Code != fakeResp.StatusCode {
		t.Errorf("Status code = '%v', want %v", w.Code, fakeResp.StatusCode)
	}
	if w.Header().Get("Foo") != fakeResp.Headers["Foo"] {
		t.Errorf("Foo header = '%v', want %v", w.Header().Get("Foo"), fakeResp.Headers["Foo"])
	}
	body, _ := ioutil.ReadAll(w.Body)
	if !bytes.Equal(body, fakeResp.Body) {
		t.Errorf("Body = '%v', want %v", string(body), string(fakeResp.Body))
	}
}
