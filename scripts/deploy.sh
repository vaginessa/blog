#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

cd go
GOOS=linux GOARCH=amd64 go build -o blog_app_linux
fab deploy
