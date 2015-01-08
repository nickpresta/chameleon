#!/bin/sh

export TEST_APP_PORT=10010
export TEST_CHAMELEON_PORT=6010

trap 'kill $(jobs -p) > /dev/null 2>&1' EXIT  # Cleanup our servers on exit

cd $(dirname $0)

chameleon -data ./testing_data -url https://httpbin.org/ -hasher="python ./hasher.py" \
    -host localhost:$TEST_CHAMELEON_PORT -verbose &
TEST_SERVICE_URL=http://localhost:$TEST_CHAMELEON_PORT/ python app.py > /dev/null 2>&1 &

sleep 3  # Let the servers spin up

python tests.py > results.txt 2>&1
TEST_RESULT=$?

cat results.txt
rm -f results.txt

exit $TEST_RESULT
