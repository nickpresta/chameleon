package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
)

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
	command *exec.Cmd
}

func (k cmdHasher) marshalRequest(r *http.Request) string {
	return `{"foo":"bar"}`
}

func (k cmdHasher) Hash(r *http.Request) string {
	return ""
}

// NewCmdHasher creates a new Hasher based on a command string
func NewCmdHasher(command string) cmdHasher {
	return cmdHasher{
		command: exec.Command("sh", "-c", command),
	}
}
