#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

cd go
#go run *.go
go build -o blog_app *.go
./blog_app
