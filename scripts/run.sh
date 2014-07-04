#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

rm -rf $TMPDIR/godep
echo "building"
godep go build -o blog_app *.go
echo "running"
./blog_app
rm blog_app
