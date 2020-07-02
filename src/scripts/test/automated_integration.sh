#!/usr/bin/env bash

IP="172.17.0.2"
PORT="8080"

FAILED_TEST="no"

# good request
RET1=$(curl -s -X POST -H "Content-Type: application/json" http://$IP:$PORT/api/dynupd -d \
    '{"username": "alex", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}')
RET1CODE=$(echo $RET1 | jq -rM .code)
if [ "$RET1CODE" != "200" ]; then
    echo "Test of good request failed."
    echo "Expected code 200 but got: $RET1CODE"
    FAILED_TEST="yes"
else
    RETADDR=$(host host1.example.com $IP | head -1 | awk '{print $4}')
    if [ "$RETADDR" != "192.0.2.4" ]; then
        FAILED_TEST="yes"
        echo "Test of host lookup on good request failed"
        echo "Expected 192.0.2.4 but got: $RETADDR"
    fi
fi

# bad password
RET2=$(curl -s -X POST -H "Content-Type: application/json" http://$IP:$PORT/api/dynupd -d \
    '{"username": "alex", "password": "badpassword", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}')
RET2CODE=$(echo $RET2 | jq -rM .code)
if [ "$RET2CODE" != "403" ]; then
    echo "Test of bad password failed."
    echo "Expected code 403 but got: $RET2CODE"
    FAILED_TEST="yes"
fi

# bad username
RET3=$(curl -s -X POST -H "Content-Type: application/json" http://$IP:$PORT/api/dynupd -d \
    '{"username": "notalex", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}')
RET3CODE=$(echo $RET3 | jq -rM .code)
if [ "$RET3CODE" != "403" ]; then
    echo "Test of bad username failed."
    echo "Expected code 403 but got: $RET3CODE"
    FAILED_TEST="yes"
fi

# bad hosts
RET4=$(curl -s -X POST -H "Content-Type: application/json" http://$IP:$PORT/api/dynupd -d \
    '{"username": "alex", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "badhost2" ], "version": "0.1.0"}')
RET4CODE=$(echo $RET4 | jq -rM .code)
if [ "$RET4CODE" != "418" ]; then
    echo "Test of bad hosts failed."
    echo "Expected code 418 but got: $RET4CODE"
    FAILED_TEST="yes"
fi

# malformed request - missing password
RET5=$(curl -s -X POST -H "Content-Type: application/json" http://$IP:$PORT/api/dynupd -d \
    '{"username": "alex", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}')
RET5CODE=$(echo $RET5 | jq -rM .code)
if [ "$RET5CODE" != "400" ]; then
    echo "Test of malformed request - missing password failed."
    echo "Expected code 400 but got: $RET5CODE"
    FAILED_TEST="yes"
fi

# malformed request - totally bogus
RET6=$(curl -s -X POST -H "Content-Type: application/json" http://$IP:$PORT/api/dynupd -d \
    '{"malformed": "data"}')
RET6CODE=$(echo $RET6 | jq -rM .code)
if [ "$RET6CODE" != "400" ]; then
    echo "Test of malformed request - totally bogus failed."
    echo "Expected code 400 but got: $RET6CODE"
    FAILED_TEST="yes"
fi

# bad method - the PUT method seems like a resonable mistake people would make...
RET7=$(curl -s -X PUT -H "Content-Type: application/json" http://$IP:$PORT/api/dynupd -d \
    '{"username": "alex", "password": "password", "ipaddr": "192.0.2.4", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}')
RET7CODE=$(echo $RET7 | jq -rM .code)
if [ "$RET7CODE" != "405" ]; then
    echo "Test of bad method failed."
    echo "Expected code 405 but got: $RET7CODE"
    FAILED_TEST="yes"
fi

# not sure how to induce an internal failure, maybe by using a different passwd file with errors in it?

if [ "$FAILED_TEST" == "no" ]; then
    echo "All tests passed!"
fi
