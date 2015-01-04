package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

// Request embeds an *http.Request to support custom JSON encoding.
type request struct {
	*http.Request
}

type requestURL struct {
	Host     string
	Path     string
	RawQuery string
	Scheme   string
}

// A Hasher interface is used to generate a key for a given request.
type Hasher interface {
	Hash(r *http.Request) string
}

type defaultHasher struct {
}

// NewHasher creates a new default Hasher
func NewHasher() defaultHasher {
	return defaultHasher{}
}

func (k defaultHasher) Hash(r *http.Request) string {
	hasher := md5.New()
	hash := r.URL.RequestURI() + r.Method
	hasher.Write([]byte(hash))

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

type cmdHasher struct {
	command string
}

func (k cmdHasher) newCommand() *exec.Cmd {
	return exec.Command("sh", "-c", k.command)
}

// MarshalJSON returns a JSON representation of a Request.
// This differs from using the built-in JSON Marshal on an *http.Request
// by embedding the body (base64 encoded), and removing fields that
// aren't important.
func (r *request) MarshalJSON() ([]byte, error) {
	var body bytes.Buffer
	body.ReadFrom(r.Body)
	bodyBytes := body.Bytes()
	r.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))

	return json.Marshal(struct {
		BodyBase64    []byte
		ContentLength int64
		Headers       http.Header
		Method        string
		URL           requestURL
	}{
		BodyBase64:    bodyBytes,
		ContentLength: r.ContentLength,
		Headers:       r.Header,
		Method:        r.Method,
		URL: requestURL{
			Host:     r.URL.Host,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
			Scheme:   r.URL.Scheme,
		},
	})
}

func (k cmdHasher) Hash(r *http.Request) string {
	command := k.newCommand()
	encodedReq, err := json.Marshal(&request{r})
	if err != nil {
		panic(err)
	}
	command.Stdin = strings.NewReader(string(encodedReq))

	var stderr bytes.Buffer
	command.Stderr = &stderr

	out, err := command.Output()
	defer command.Process.Kill()
	if err != nil {
		log.Printf("%v:\nSTDOUT:\n%v\n\nSTDERR:\n%v", err, string(out), stderr.String())
		panic(err)
	}

	hasher := md5.New()
	hasher.Write(out)
	return hex.EncodeToString(hasher.Sum(nil))
}

// NewCmdHasher creates a new Hasher based on a command string
func NewCmdHasher(command string) cmdHasher {
	return cmdHasher{
		command: command,
	}
}
