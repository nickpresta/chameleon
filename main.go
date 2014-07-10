package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/NickPresta/gomeleon/cache"
	"github.com/NickPresta/gomeleon/config"
)

var configPath *string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get current working directory")
	}
	configFile := path.Join(cwd, "config.json")
	configPath = flag.String("config", configFile, "Full path to configuration file")
}

func cachedProxyMiddleware(handler http.HandlerFunc, server config.ServerDefinition, c cache.Cacher) http.HandlerFunc {
	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Change the host for the request for this server config
		r.Host = parsedURL.Host
		r.URL.Host = r.Host
		r.URL.Scheme = parsedURL.Scheme
		r.RequestURI = ""

		log.Printf("-> Proxying to %s\n", r.URL)

		key := c.Key(r)
		response := c.Get(key)

		// We don't have a cached response yet
		if response == nil {
			// Create a recorder, so we can get data out and modify it (if needed)
			rec := httptest.NewRecorder()
			handler(w, r) // Actually call our handler
			copyHeaders(w.Header(), rec.Header())
			io.Copy(w, rec.Body) // Write out response

			c.Put(key, rec)

			return
		}

		// Fetch from cache, return that response

	}
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) // Proxy through
}

func main() {
	flag.Parse()
	config, err := config.ParseConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	for _, server := range config {
		wg.Add(1)
		go func() {
			defer wg.Done()

			cacher := cache.NewDiskCacher(server.DataDirectory)

			http.HandleFunc("/", cachedProxyMiddleware(proxyHandler, server, cacher))
			http.ListenAndServe(fmt.Sprintf(":%d", server.Port), nil)
		}()
	}
	wg.Wait()
}
