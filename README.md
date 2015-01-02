# What is chameleon?

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

* Check out the [example](./example) directory for a small app that uses chameleon to create reliable tests.

To run chameleon, you can:

    chameleon -data ./httpstatus -url http://httpstat.us -verbose

The directory `httpstatus` must already exist before running.

See `chameleon -help` for more information.

### How chameleon caches responses

chameleon makes a key hash for a given request URI and method and uses that to cache content. What that means:

* a request of `GET /foo/` will be cached differently than `GET /bar/`
* a request of `GET /foo/5` will be cached differently than `GET /foo/6`
* a request of `DELETE /foo/5` will be cached differently than `DELETE /foo/6`
* a request of `POST /foo` with a body of `{"hi":"hello}` will be cached the same as a
  request of `POST /foo` with a body of `{"spam":"eggs"}`. To get around this, set a header of `chameleon-hash-body`
  to any value. This will instruct chameleon to use the entire body as part of the key hash.

