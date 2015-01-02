# -*- coding: utf-8 -*-

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


if __name__ == '__main__':
    unittest.main()
