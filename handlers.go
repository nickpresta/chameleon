package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// CachedProxyMiddleware proxies a given URL and stores/fetches content from a Cacher
func CachedProxyMiddleware(handler http.HandlerFunc, serverURL *url.URL, c Cacher, h Hasher) http.HandlerFunc {
	parsedURL, err := url.Parse(serverURL.String())
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Change the host for the request for this configuration
		r.Host = parsedURL.Host
		r.URL.Host = r.Host
		r.URL.Scheme = parsedURL.Scheme
		r.RequestURI = ""

		hash := h.Hash(r)
		response := c.Get(hash)

		if response != nil {
			log.Printf("-> Proxying [cached: %v] to %v\n", hash, r.URL)
		} else {
			// We don't have a cached response yet
			log.Printf("-> Proxying [not cached: %v] to %v\n", hash, r.URL)

			// Create a recorder, so we can get data out and modify it (if needed)
			rec := httptest.NewRecorder()
			handler(rec, r) // Actually call our handler

			response = c.Put(hash, rec)
		}

		for k, v := range response.Headers {
			w.Header().Add(k, v)
		}
		w.WriteHeader(response.StatusCode)
		// If this fails, there isn't much to do
		_, _ = io.Copy(w, bytes.NewReader(response.Body))
	}
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// ProxyHandler implements a standard HTTP handler to proxy a given request and returns the response
func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		// If this fails, there isn't much to do
		_ = resp.Body.Close()
	}()
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	// If this fails, there isn't much to do
	_, _ = io.Copy(w, resp.Body) // Proxy through
}
