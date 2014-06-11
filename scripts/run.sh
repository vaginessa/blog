#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

#go run *.go
go build -o blog_app *.go
./blog_app
rm blog_app
