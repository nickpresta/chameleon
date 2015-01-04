# -*- coding: utf-8 -*-

import base64
import json
import sys


def hasher(request):
    out = request['Method'] + request['URL']['Path']
    if request['Headers'].get('Chameleon-Hash-Body', [''])[0] == 'true':
        out += base64.b64decode(request['BodyBase64'])
    return out


def main(stdin):
    request = json.loads(sys.stdin.read())
    out = hasher(request)
    sys.stdout.write(out)


if __name__ == '__main__':
    main(sys.stdin)
