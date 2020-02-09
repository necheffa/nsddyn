#!/usr/bin/env bash

curl -X POST -H "Content-Type: application/json" http://127.0.0.1:8080/api/dynupd -d \
    '{"username": "alex", "password": "password", "ipaddr": "127.0.0.100", "hostnames": [ "myhost1", "myhost2" ]}'
