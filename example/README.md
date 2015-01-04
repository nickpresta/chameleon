# chameleon Example

This example illustrates how to setup chameleon with a custom hasher and how you might integrate it into your application.

## Application

This application's purpose to is accept a given HTTP status code (like `200`) and return the text associated with
that status code (`OK` in the `200` case).

The app in this directory calls [https://httpbin.org](https://httpbin.org) to fetch a response for a given status code,
grabs the "message" for that status code (`OK`, `I'M A TEAPOT`, etc) and returns that in the response to the user.

To run this application (you need Python 2.x):

        $ TEST_PORT=10005 python app.py

Then use cURL, your browser, etc, and issue an HTTP GET request to `localhost:10005/418`. You should see `I'M A TEAPOT`
as the response body.

## Testing this service

There are some accompanying user tests (E2E, API, what ever you call them) in the file `tests.py`. Run it like so:

        $ TEST_PORT=10005 python tests.py

You should see a bunch of unit tests pass that look like this (note the time it takes):

        $ TEST_PORT=10005 python tests.py
        ........
        ----------------------------------------------------------------------
        Ran 8 tests in 1.970s

        OK

You could imagine tests that check JSON error payloads conform to a certain structure, that response headers are present,
and a whole list of other things you care about in an end-to-end test scenario.

## Applicability

Imagine you are writing an app that depended on an external service to do its job.

What would you do if your external service was rate limiting you? How about only allowing access from specific
IP addresses? What if the external service was slow?

You could proxy and cache the backend service and allow your E2E tests to behave normally and with real, valid data.

## How to integrate chameleon

This assumes you're running chameleon from this `example` directory.

1. Set up chameleon to proxy calls to https://httpbin.org:

        $ mkdir httpbin
        $ chameleon -data ./httpbin -port 6005 -verbose -url https://httpbin.org/ -hasher 'python ./hasher.py'

1. Instruct our application to use chameleon to make requests. We set the `TEST_SERVICE_URL` to chameleon:

        $ TEST_PORT=10005 TEST_SERVICE_URL=http://localhost:6005/ python app.py

1. Run our tests again:


        $ TEST_PORT=10005 python tests.py
        ........
        ----------------------------------------------------------------------
        Ran 8 tests in 1.728s

        OK

You will notice that our test run isn't much faster. If you flip over to your view of chameleon, you will see:

        Starting proxy for 'https://httpbin.org/'
        -> Proxying [not cached: 116f933e981e92c994619116ee37fd30] to https://httpbin.org/status/200
        -> Proxying [not cached: 7c68a3b062b22caf6b2ca517027611bc] to https://httpbin.org/status/418
        -> Proxying [not cached: f6970b3f15df6d952f33387988c04967] to https://httpbin.org/status/500
        -> Proxying [cached: 116f933e981e92c994619116ee37fd30] to https://httpbin.org/status/200

We can see that chameleon actually hit `https://httpbin.org/status/:code` three times and then the fourth time,
it had a cache for the `200` code so it returned the cached version.

If we run our tests again, we see:

        $ TEST_PORT=10005 python tests.py
        ........
        ----------------------------------------------------------------------
        Ran 8 tests in 0.015s

        OK

In all four cases, chameleon returned the responses from disk. This resulted in a much faster test run,
and if our backend service started to throttle us, or we wanted to run these tests from somewhere that couldn't
reach httpbin, we still could.

## Conclusions

It can be fairly trivial to integrate chameleon into your testing workflow. In fact, this example was the most
complicated example of running chameleon. For simple services, you may not need a custom hasher, in which case the default
hasher does the Right Thing™ (as described in the docs).
