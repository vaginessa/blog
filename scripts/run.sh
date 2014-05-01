#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

cd go
#go run *.go
go build -o blog_app *.go || exit 1
./blog_app
