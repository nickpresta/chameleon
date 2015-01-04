# chameleon

[![Build Status](https://img.shields.io/travis/NickPresta/chameleon.svg?style=flat)](https://travis-ci.org/NickPresta/chameleon)
[![License](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](https://tldrlegal.com/license/mit-license)

## What is chameleon?

chameleon is a caching reverse proxy.

chameleon supports recording and replaying requests for multiple, simultaneous services.

## Why is chameleon useful?

* Proxy rate-limited APIs for local development
* Create reliable test environments
* Test services in places you normally couldn't due to firewalling, etc (CI servers being common)
* Improve speed of "interface tests" by never leaving your local network

## What can't I do with chameleon?

* Have tests that exercise a given service **right now** as results are cached
* Total control on how things are cached, frequency, rate-limiting, etc (pull requests are welcome, though!)

## How to get chameleon?

chameleon has **no** runtime dependencies. You can download a
[prebuilt binary](https://github.com/NickPresta/chameleon/releases) for your platform.

If you have Go installed, you may `go get github.com/NickPresta/chameleon` to download it to your `$GOPATH` directory.

## How to use chameleon

* Check out the [example](./example) directory for a small app that uses chameleon to create reliable tests with a custom hasher.

To run chameleon, you can:

    chameleon -data ./httpstatus -url http://httpstat.us -verbose

The directory `httpstatus` must already exist before running.

See `chameleon -help` for more information.

### How chameleon caches responses

chameleon makes a hash for a given request URI and method and uses that to cache content. What that means:

* a request of `GET /foo/` will be cached differently than `GET /bar/`
* a request of `GET /foo/5` will be cached differently than `GET /foo/6`
* a request of `DELETE /foo/5` will be cached differently than `DELETE /foo/6`
* a request of `POST /foo` with a body of `{"hi":"hello}` will be cached the same as a
  request of `POST /foo` with a body of `{"spam":"eggs"}`. To get around this, set a header of `chameleon-hash-body`
  to any value. This will instruct chameleon to use the entire body as part of the hash.

### Writing custom hasher

You can specify a custom hasher, which could be any program in any language, to determine what makes a request unique.

chameleon will communicate with this program via STDIN/STDOUT and feed the hasher a serialized `Request` (see below).
You are then responsible for returning data to chameleon to be used for that given request (which will be hashed).

This feature is especially useful if you have to cache content based on the body of a request
(XML payload, specific keys in JSON payload, etc).

See the [example hasher](./example/hasher.py) for a sample hasher that emulates the default hasher.

#### Structure of Request

Below is an example Request serialized to JSON.

```json
{
    "BodyBase64":"eyJmb28iOiAiYmFyIn0=",
    "ContentLength":14,
    "Headers":{
        "Accept":[
            "application/json"
        ],
        "Accept-Encoding":[
            "gzip, deflate"
        ],
        "Authorization":[
            "Basic dXNlcjpwYXNzd29yZA=="
        ],
        "Connection":[
            "keep-alive"
        ],
        "Content-Length":[
            "14"
        ],
        "Content-Type":[
            "application/json; charset=utf-8"
        ],
        "User-Agent":[
            "HTTPie/0.7.2"
        ]
    },
    "Method":"POST",
    "URL":{
        "Host":"httpbin.org",
        "Path":"/post",
        "RawQuery":"q=search+term%23home",
        "Scheme":"https"
    }
}
```

Field | Description
----- | -----------
BodyBase64 | Body is the request's body, base64 encoded
ContentLength | ContentLength records the length of the associated content after being base64 decoded
Headers | Headers is a map of request lines to value lists. HTTP defines that header names are case-insensitive. Header names have been canonicalized, making the first character and any characters following a hyphen uppercase and the rest lowercase.
Method | Method specifies the HTTP method (`GET`, `POST`, `PUT`, etc.)
URL | URL is an object containing `Host`, the HTTP Host in the form of 'host' or 'host:port', `Path`, the request path including trailing slash, `RawQuery`, encoded query string values without '?', and `Scheme`, the URL scheme 'http', 'https'

## Getting help

Please [open an issue](https://github.com/nickpresta/chameleon/issues) for any bugs encountered, features requests, or
general troubleshooting.

## Authors

[Nick Presta](http://nickpresta.ca) ([@NickPresta](https://twitter.com/NickPresta))

## License

Please see [LICENSE](./LICENSE)
