#!/usr/bin/env bash

curl -X POST -H "Content-Type: application/json" http://127.0.0.1:5000/api/dynupd -d \
    '{"username": "alex", "password": "password", "ipaddr": "127.0.0.100", "hostname": "myhost"}'

