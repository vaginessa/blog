#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

rm -rf $TMPDIR/godep
godep go build -o blog_app *.go
rm blog_app
