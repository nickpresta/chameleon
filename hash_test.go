package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"testing"
)

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

func TestDefaultHasherWithBody(t *testing.T) {
	hasher := DefaultHasher{}

	body := "HASH THIS BODY"
	req, _ := http.NewRequest("POST", "/foobar", strings.NewReader(body))
	req.Header.Set("chameleon-hash-body", "true")
	hash := hasher.Hash(req)

	md5Hasher := md5.New()
	md5Hasher.Write([]byte(req.URL.RequestURI() + req.Method + body))
	expected := hex.EncodeToString(md5Hasher.Sum(nil))
	if hash != expected {
		t.Errorf("Got: `%v`; Expected: `%v`", hash, expected)
	}
}

func TestCmdHasher(t *testing.T) {
	var stdin bytes.Buffer
	hasher := CmdHasher{Command: "/bin/cat", Commander: testCommander{stdin: &stdin}}
	req, _ := http.NewRequest("POST", "/foobar", strings.NewReader("HASH THIS BODY"))
	req.Header.Set("chameleon-hash-body", "true")
	hash := hasher.Hash(req)

	md5Hasher := md5.New()
	// our command just echoes back what we gave it, so all of stdin should be included in the hash
	md5Hasher.Write(stdin.Bytes())
	expected := hex.EncodeToString(md5Hasher.Sum(nil))
	if hash != expected {
		t.Errorf("Got: `%v`; Expected: `%v`", hash, expected)
	}
}
