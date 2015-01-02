# -*- coding: utf-8 -*-

import BaseHTTPServer
import os
import urllib2
import urlparse

SERVICE_URL = os.getenv('TEST_SERVICE_URL', 'https://httpbin.org/')
PORT = int(os.getenv('TEST_PORT', 9001))

STATUS_SERVICE_URL = urlparse.urljoin(SERVICE_URL, '/status/')


class MyHandler(BaseHTTPServer.BaseHTTPRequestHandler):

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


def main():
    print('Serving on port {}'.format(PORT))
    server = BaseHTTPServer.HTTPServer(('localhost', PORT), MyHandler)
    server.serve_forever()


if __name__ == '__main__':
    main()
