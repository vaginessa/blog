#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

rm -rf $TMPDIR/godep
GOOS=linux GOARCH=amd64 gdep go build -o blog_app_linux
fab deploy
rm blog_app_linux
