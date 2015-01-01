package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/NickPresta/chameleon/cache"
	"github.com/NickPresta/chameleon/config"
)

func CachedProxyMiddleware(handler http.HandlerFunc, server config.ServerDefinition, c cache.Cacher) http.HandlerFunc {
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

		key := c.Key(r)
		response := c.Get(key)

		cached := "not cached"
		if response != nil {
			cached = "cached"
		}

		log.Printf("-> Proxying [%v] to %v\n", cached, r.URL)

		// We don't have a cached response yet
		if response == nil {
			// Create a recorder, so we can get data out and modify it (if needed)
			rec := httptest.NewRecorder()
			handler(rec, r) // Actually call our handler

			c.Put(key, rec)

			copyHeaders(w.Header(), rec.Header())
			w.WriteHeader(rec.Code)
			io.Copy(w, rec.Body) // Write out response

			return
		}

		// Fetch from cache, return that response
		for k, v := range response.Headers {
			w.Header().Add(k, v)
		}
		w.WriteHeader(response.StatusCode)
		io.Copy(w, bytes.NewReader(response.Body))
	}
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
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
