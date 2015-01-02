# -*- coding: utf-8 -*-

import json
import os
import unittest
import urllib2

PORT = int(os.getenv('TEST_PORT', 9001))


def get_name(code):
    url = 'http://localhost:{}/{}'.format(PORT, code)
    try:
        resp = urllib2.urlopen(url)
    except urllib2.HTTPError as exc:
        resp = exc
    return resp.read()


class MyTest(unittest.TestCase):

    def test_200_returns_ok(self):
        self.assertEqual('OK', get_name(200))

    def test_418_returns_teapot(self):
        self.assertEqual("I'M A TEAPOT", get_name(418))

    def test_500_internal_server_error(self):
        self.assertEqual('INTERNAL SERVER ERROR', get_name(500))

    def test_content_type_is_text_plain(self):
        url = 'http://localhost:{}/200'.format(PORT)
        resp = urllib2.urlopen(url)
        self.assertEqual('text/plain', resp.headers['content-type'])

    def test_post_returns_post_body(self):
        url = 'http://localhost:{}/post'.format(PORT)
        req = urllib2.Request(url, json.dumps({'foo': 'bar'}), {'Content-type': 'application/json'})
        req.get_method = lambda: 'POST'
        resp = urllib2.urlopen(req)
        parsed = json.loads(resp.read())
        self.assertEqual({'foo': 'bar'}, parsed['json'])

        # now with hashed
        url = 'http://localhost:{}/post_with_body'.format(PORT)
        req = urllib2.Request(url, json.dumps({'post': 'body'}), {'Content-type': 'application/json'})
        req.get_method = lambda: 'HASHED'
        resp = urllib2.urlopen(req)
        parsed = json.loads(resp.read())
        self.assertEqual({'post': 'body'}, parsed['json'])

    def test_patch_returns_body(self):
        url = 'http://localhost:{}/patch'.format(PORT)
        req = urllib2.Request(url, json.dumps({'hi': 'hello'}), {'Content-type': 'application/json'})
        req.get_method = lambda: 'PATCH'
        resp = urllib2.urlopen(req)
        parsed = json.loads(resp.read())
        self.assertEqual({'hi': 'hello'}, parsed['json'])

    def test_put_returns_body(self):
        url = 'http://localhost:{}/put'.format(PORT)
        req = urllib2.Request(url, json.dumps({'spam': 'eggs'}), {'Content-type': 'application/json'})
        req.get_method = lambda: 'PUT'
        resp = urllib2.urlopen(req)
        parsed = json.loads(resp.read())
        self.assertEqual({'spam': 'eggs'}, parsed['json'])

    def test_delete_returns_200(self):
        url = 'http://localhost:{}/delete'.format(PORT)
        req = urllib2.Request(url)
        req.get_method = lambda: 'DELETE'
        resp = urllib2.urlopen(req)
        self.assertEqual(200, resp.getcode())


if __name__ == '__main__':
    unittest.main()
