#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

GOOS=linux GOARCH=amd64 godep go build -o blog_app_linux
fab deploy
rm blog_app_linux
