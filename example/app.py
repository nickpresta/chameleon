# -*- coding: utf-8 -*-

import BaseHTTPServer
import os
import urllib2
import urlparse

SERVICE_URL = os.getenv('TEST_SERVICE_URL', 'https://httpbin.org/')
TEST_APP_PORT = int(os.getenv('TEST_APP_PORT', 9001))

STATUS_SERVICE_URL = urlparse.urljoin(SERVICE_URL, '/status/')
POST_SERVICE_URL = urlparse.urljoin(SERVICE_URL, '/post')
PUT_SERVICE_URL = urlparse.urljoin(SERVICE_URL, '/put')
PATCH_SERVICE_URL = urlparse.urljoin(SERVICE_URL, '/patch')
DELETE_SERVICE_URL = urlparse.urljoin(SERVICE_URL, '/delete')


class MyHandler(BaseHTTPServer.BaseHTTPRequestHandler):

    def _do_patch_post_put(self, url, method, headers=None):
        if headers is None:
            headers = {}
        headers.update({'Content-Type': 'application/json'})
        content_len = int(self.headers.getheader('content-length', 0))
        body = self.rfile.read(content_len)

        req = urllib2.Request(url, body, headers)
        req.get_method = lambda: method
        try:
            resp = urllib2.urlopen(req)
        except urllib2.HTTPError as exc:
            resp = exc

        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(resp.read())

    def do_PATCH(self):
        self._do_patch_post_put(PATCH_SERVICE_URL, 'PATCH')

    def do_PUT(self):
        self._do_patch_post_put(PUT_SERVICE_URL, 'PUT')

    def do_POST(self):
        self._do_patch_post_put(POST_SERVICE_URL, 'POST')

    def do_DELETE(self):
        req = urllib2.Request(DELETE_SERVICE_URL)
        req.get_method = lambda: 'DELETE'
        try:
            resp = urllib2.urlopen(req)
        except urllib2.HTTPError as exc:
            resp = exc

        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        self.wfile.write(resp.read())

    def do_GET(self):
        # requests to /200 will forward the request to STATUS_SERVICE_URL/200, etc
        # and return a response with the status code text string
        url = urlparse.urljoin(STATUS_SERVICE_URL, self.path[1:])
        try:
            resp = urllib2.urlopen(url)
        except urllib2.HTTPError as exc:
            resp = exc
        self.send_response(200)
        self.send_header('Content-type', 'text/plain')
        self.end_headers()
        self.wfile.write(resp.msg.upper())

    def do_HASHED(self):
        # Custom method that doesn't hash a post with body
        self._do_patch_post_put(POST_SERVICE_URL, 'POST', {'chameleon-no-hash-body': 'true'})

    def do_SEEDED(self):
        url = urlparse.urljoin(SERVICE_URL, self.path[1:])
        try:
            resp = urllib2.urlopen(url)
        except urllib2.HTTPError as exc:
            resp = exc
        self.send_response(resp.getcode())
        self.send_header('Content-type', resp.headers['content-type'])
        self.end_headers()
        self.wfile.write(resp.read())

    def do_REQUESTHASH(self):
        content_len = int(self.headers.getheader('content-length', 0))
        body = self.rfile.read(content_len)

        req = urllib2.Request(POST_SERVICE_URL, body, self.headers)
        req.get_method = lambda: 'POST'
        try:
            resp = urllib2.urlopen(req)
        except urllib2.HTTPError as exc:
            resp = exc

        self.send_response(200)
        for k, v in resp.headers.dict.viewitems():
            self.send_header(k, v)
        self.end_headers()
        self.wfile.write(resp.read())


def main():
    print('Serving on port {}'.format(TEST_APP_PORT))
    server = BaseHTTPServer.HTTPServer(('localhost', TEST_APP_PORT), MyHandler)
    server.serve_forever()


if __name__ == '__main__':
    main()
