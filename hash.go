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

// DefaultHasher is the default implementation of a Hasher
type DefaultHasher struct {
}

// NewHasher creates a new default Hasher
func NewHasher() DefaultHasher {
	return DefaultHasher{}
}

// Hash returns a hash for a given request.
// The default behavior is to hash the URL and method
// but if the header 'chameleon-hash-body' exists, the body
// will be used to hash as well.
func (k DefaultHasher) Hash(r *http.Request) string {
	hasher := md5.New()
	hash := r.URL.RequestURI() + r.Method
	// This method always succeeds
	_, _ = hasher.Write([]byte(hash))

	if r.Header.Get("chameleon-hash-body") != "" {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			panic(err)
		}
		bufBytes := buf.Bytes()

		_, err = io.Copy(hasher, bytes.NewReader(bufBytes))
		if err != nil {
			panic(err)
		}
		// Put the body back on the request so it can read again
		r.Body = ioutil.NopCloser(bytes.NewReader(bufBytes))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// CmdHasher is an implementation of a Hasher which uses other commands to generate a hash via STDIN/STDOUT.
type CmdHasher struct {
	command string
}

func (k CmdHasher) newCommand() *exec.Cmd {
	return exec.Command("sh", "-c", k.command)
}

// MarshalJSON returns a JSON representation of a Request.
// This differs from using the built-in JSON Marshal on an *http.Request
// by embedding the body (base64 encoded), and removing fields that
// aren't important.
func (r *request) MarshalJSON() ([]byte, error) {
	var body bytes.Buffer
	_, err := body.ReadFrom(r.Body)
	if err != nil {
		return nil, err
	}
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

// Hash returns a hash for a given request.
// This implementation defers to an external command for a hash and communicates via STDIN/STDOUT.
func (k CmdHasher) Hash(r *http.Request) string {
	command := k.newCommand()
	encodedReq, err := json.Marshal(&request{r})
	if err != nil {
		panic(err)
	}
	command.Stdin = strings.NewReader(string(encodedReq))

	var stderr bytes.Buffer
	command.Stderr = &stderr

	out, err := command.Output()
	defer func() {
		// If this fails, there isn't much to do
		_ = command.Process.Kill()
	}()
	if err != nil {
		log.Printf("%v:\nSTDOUT:\n%v\n\nSTDERR:\n%v", err, string(out), stderr.String())
		panic(err)
	}

	hasher := md5.New()
	// This method always succeeds
	_, _ = hasher.Write(out)
	return hex.EncodeToString(hasher.Sum(nil))
}

// NewCmdHasher creates a new Hasher based on a command string
func NewCmdHasher(command string) CmdHasher {
	return CmdHasher{
		command: command,
	}
}
