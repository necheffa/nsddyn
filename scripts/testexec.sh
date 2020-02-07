#!/usr/bin/env bash

python3-coverage run -m pytest
python3-coverage report
python3-coverage html
