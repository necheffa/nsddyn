#!/usr/bin/env bash

# good request
echo "Good request: "
curl -X POST -H "Content-Type: application/json" http://127.0.0.1:8080/api/dynupd -d \
    '{"username": "alex", "password": "password", "ipaddr": "127.0.0.100", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}'

echo ""

# bad password
echo "Bad password: "
curl -X POST -H "Content-Type: application/json" http://127.0.0.1:8080/api/dynupd -d \
    '{"username": "alex", "password": "badpassword", "ipaddr": "127.0.0.100", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}'

echo ""

# bad username
echo "Bad username: "
curl -X POST -H "Content-Type: application/json" http://127.0.0.1:8080/api/dynupd -d \
    '{"username": "notalex", "password": "password", "ipaddr": "127.0.0.100", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}'

echo ""

# bad hosts
echo "Bad host: "
curl -X POST -H "Content-Type: application/json" http://127.0.0.1:8080/api/dynupd -d \
    '{"username": "alex", "password": "password", "ipaddr": "127.0.0.100", "hostnames": [ "host1", "badhost2" ], "version": "0.1.0"}'

echo ""

# malformed request - missing password
echo "Missing password: "
curl -X POST -H "Content-Type: application/json" http://127.0.0.1:8080/api/dynupd -d \
    '{"username": "alex", "ipaddr": "127.0.0.100", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}'

echo ""

# malformed request - totally bogus
echo "Totally malformed: "
curl -X POST -H "Content-Type: application/json" http://127.0.0.1:8080/api/dynupd -d \
    '{"malformed": "data"}'

echo ""

# bad method - the PUT method seems like a resonable mistake people would make...
echo "PUT method instead of POST: "
curl -X PUT -H "Content-Type: application/json" http://127.0.0.1:8080/api/dynupd -d \
    '{"username": "alex", "password": "password", "ipaddr": "127.0.0.100", "hostnames": [ "host1", "host2" ], "version": "0.1.0"}'

echo ""

# not sure how to induce an internal failure, maybe by using a different passwd file with errors in it?
