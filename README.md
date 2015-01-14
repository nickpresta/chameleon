# chameleon

[![Build Status](https://img.shields.io/travis/NickPresta/chameleon/master.svg?style=flat)](https://travis-ci.org/NickPresta/chameleon)
[![Coveralls](https://img.shields.io/coveralls/NickPresta/chameleon/master.svg?style=flat)](https://coveralls.io/r/NickPresta/chameleon)
[![License](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](https://tldrlegal.com/license/mit-license)

## What is chameleon?

chameleon is a caching reverse proxy.

chameleon supports recording and replaying requests with the ability to customize how responses are stored.

## Why is chameleon useful?

* Proxy rate-limited APIs for local development
* Create reliable test environments
* Test services in places you normally couldn't due to firewalling, etc (CI servers being common)
* Improve speed of tests by never leaving your local network
* Inspect recorded APIs responses for exploratory testing
* Stub out unimplemented API endpoints during development

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

    chameleon -data ./httpbin -url http://httpbin.org -verbose

The directory `httpbin` must already exist before running.

See `chameleon -help` for more information.

### Preseeding the cache

If you want to configure the cache at runtime without having to depend on an external service, you may preseed the cache
via HTTP. This is particularly useful for mocking out services which don't yet exist.

To preseed a request, issue a JSON `POST` request to chameleon at the `_seed` endpoint with the following payload:

Field | Description
----- | -----------
`Request` | Request is the request payload including a URL, Method and Body
`Response` | Response is the response to be cached and sent back for a given request

**Request**

Field | Description
----- | -----------
`Body` | Body is the content for the request. May be empty where body doesn't make sense (e.g. `GET` requests)
`Method` | Method is the HTTP method used to match the incoming request. Case insensitive, supports arbitrary methods
`URL` | URL is the absolute or relative URL to match in requests. Only the path and querystring are used

**Response**

Field | Description
----- | -----------
`Body` | Body is the content for the request. May be empty where body doesn't make sense (e.g. `GET` requests)
`Headers` | Headers is a map of headers in the format of string key to string value
`StatusCode` | StatusCode is the [HTTP status code](http://en.wikipedia.org/wiki/List_of_HTTP_status_codes) of the response

Repeated, duplicate requests to preseed the cache will be discarded and the cache unaffected.

Successful new preseed requests will return an `HTTP 201 CREATED` on success or `HTTP 500 INTERNAL SERVER ERROR`.
Duplicate preseed requests will return an `HTTP 200 OK` on success or `HTTP 500 INTERNAL SERVER ERROR` on failure.

Here is an example of preseeding the cache with a JSON response for a `GET` request for `/foobar`.

```python
import requests

preseed = json.dumps({
    'Request': {
        'Body': '',
        'URL': '/foobar',
        'Method': 'GET',
    },
    'Response': {
        'Body': '{"key": "value"}',
        'Headers': {
            'Content-Type': 'application/json',
            'Other-Header': 'something-else',
        },
        'StatusCode': 200,
    },
})

response = requests.post('http://localhost:6005/_seed', data=preseed)
if response.status_code in (200, 201):
    # Created, or duplicate
else:
    # Error, print it out
    print(response.content)

# Continue tests as normal
# Making requests to `/foobar` will return `{"key": "value"}`
# without hitting the proxied service
```

Check out the [example](./example) directory to see preseeding in action.

### How chameleon caches responses

chameleon makes a hash for a given request URI, request method and request body and uses that to cache content. What that means:

* a request of `GET /foo/` will be cached differently than `GET /bar/`
* a request of `GET /foo/5` will be cached differently than `GET /foo/6`
* a request of `DELETE /foo/5` will be cached differently than `DELETE /foo/6`
* a request of `POST /foo` with a body of `{"hi":"hello}` will be cached differently than a request of `POST /foo` with a body of `{"spam":"eggs"}`. To ignore the request body, set a header of `chameleon-no-hash-body` to any value. This will instruct chameleon to ignore the body as part of the hash.

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
Headers | Headers is a map of request lines to value lists. HTTP defines that header names are case-insensitive. Header names have been canonicalized, making the first character and any characters following a hyphen uppercase and the rest lowercase
Method | Method specifies the HTTP method (`GET`, `POST`, `PUT`, etc.)
URL | URL is an object containing `Host`, the HTTP Host in the form of 'host' or 'host:port', `Path`, the request path including trailing slash, `RawQuery`, encoded query string values without '?', and `Scheme`, the URL scheme 'http', 'https'

## Getting help

Please [open an issue](https://github.com/nickpresta/chameleon/issues) for any bugs encountered, features requests, or
general troubleshooting.

## Authors

[Nick Presta](http://nickpresta.ca) ([@NickPresta](https://twitter.com/NickPresta))

Thanks to [@mdibernardo](https://twitter.com/mdibernardo) for the inspiration.

## License

Please see [LICENSE](./LICENSE)
