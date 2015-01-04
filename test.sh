#!/bin/sh -x

TEST_PORT=10010
CHAMELEON_PORT=6010

trap 'kill $(jobs -p)' EXIT  # Cleanup our servers on exit

chameleon -data ./testing_data -url https://httpbin.org/ -hasher="python ./example/hasher.py" -host localhost:$CHAMELEON_PORT > /dev/null 2>&1 &
TEST_PORT=$TEST_PORT TEST_SERVICE_URL=http://localhost:$CHAMELEON_PORT/ python example/app.py > /dev/null 2>&1 &

sleep 3  # Let the servers spin up

TEST_PORT=$TEST_PORT python example/tests.py
exit $?
