package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

type preseedResponse struct {
	Request struct {
		Body   string
		URL    string
		Method string
	}
	Response struct {
		Body       string
		StatusCode int
		Headers    map[string]string
	}
}

// PreseedHandler preseeds a Cacher, according to a Hasher
func PreseedHandler(cacher Cacher, hasher Hasher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		var preseedResp preseedResponse
		err := dec.Decode(&preseedResp)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err)
			return
		}

		fakeReq, err := http.NewRequest(
			preseedResp.Request.Method,
			preseedResp.Request.URL,
			strings.NewReader(preseedResp.Request.Body),
		)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err)
			return
		}
		hash := hasher.Hash(fakeReq)
		response := cacher.Get(hash)

		w.Header().Add("chameleon-request-hash", hash)
		if response != nil {
			log.Printf("-> Proxying [preseeding;cached: %v] to %v\n", hash, preseedResp.Request.URL)
			w.WriteHeader(200)
			return
		}

		log.Printf("-> Proxying [preseeding;not cached: %v] to %v\n", hash, preseedResp.Request.URL)

		rec := httptest.NewRecorder()
		rec.Body = bytes.NewBufferString(preseedResp.Response.Body)
		rec.Code = preseedResp.Response.StatusCode
		for name, value := range preseedResp.Response.Headers {
			rec.Header().Set(name, value)
		}

		// Signal to the cacher to skip the disk
		rec.Header().Set("_chameleon-seeded-skip-disk", "true")

		// Don't need the response
		_ = cacher.Put(hash, rec)
		w.WriteHeader(201)
	}
}

// CachedProxyHandler proxies a given URL and stores/fetches content from a Cacher, according to a Hasher
func CachedProxyHandler(serverURL *url.URL, cacher Cacher, hasher Hasher, proxier http.HandlerFunc) http.HandlerFunc {
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

		hash := r.Header.Get("chameleon-request-hash")
		if hash == "" {
			hash = hasher.Hash(r)
		}
		response := cacher.Get(hash)

		if response != nil {
			log.Printf("-> Proxying [cached: %v] to %v\n", hash, r.URL)
		} else {
			// We don't have a cached response yet
			log.Printf("-> Proxying [not cached: %v] to %v\n", hash, r.URL)

			// Create a recorder, so we can get data out and modify it (if needed)
			rec := httptest.NewRecorder()
			proxier(rec, r) // Actually call our handler

			response = cacher.Put(hash, rec)
		}

		for k, v := range response.Headers {
			w.Header().Add(k, v)
		}
		w.Header().Add("chameleon-request-hash", hash)
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
func ProxyHandler(skipverify bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipverify},
		}
		client := &http.Client{Transport: tr}
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
}
