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
	verbose    = flag.Bool("verbose", false, "Turn on verbose logging")
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

	log.Printf("Starting proxy for '%v'\n", serverURL.String())
	cacher := NewDiskCacher(*dataDir)
	mux := http.NewServeMux()
	mux.Handle("/", CachedProxyMiddleware(ProxyHandler, serverURL, cacher))
	log.Fatal(http.ListenAndServe(*host, mux))
}
