// +build !testing

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
)

var (
	proxiedURL = flag.String("url", "", "Fully qualified, absolute URL to proxy (e.g. https://example.com)")
	dataDir    = flag.String("data", "", "Path to a directory in which to hold the responses for this url")
	host       = flag.String("host", "localhost:6005", "Host/port on which to bind")
	cHasher    = flag.String("hasher", "", "Custom hasher program for all requests (e.g. python ./hasher.py)")
	verbose    = flag.Bool("verbose", false, "Turn on verbose logging")
	skipverify = flag.Bool("insecure-skip-verify", false, "Skips verification of server's certificate")
)

func main() {
	flag.Parse()
	if *proxiedURL == "" || *dataDir == "" {
		flag.Usage()
		os.Exit(-1)
	}

	serverURL, err := url.Parse(*proxiedURL)
	if err != nil {
		log.Fatal(err)
	}

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Printf("Starting proxy for '%v' on %v\n", serverURL.String(), *host)
	var hasher Hasher
	if *cHasher != "" {
		hasher = CmdHasher{Command: *cHasher, Commander: DefaultCommander{}}
	} else {
		hasher = DefaultHasher{}
	}
	cacher := NewDiskCacher(*dataDir)
	cacher.SeedCache()
	mux := http.NewServeMux()
	mux.Handle("/_seed", PreseedHandler(cacher, hasher))
	mux.Handle("/", CachedProxyHandler(serverURL, cacher, hasher, ProxyHandler(*skipverify)))
	log.Fatal(http.ListenAndServe(*host, mux))
}
