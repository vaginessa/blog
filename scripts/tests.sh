#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

# on the server the hierarchy is different
if [ -e go ]; then cd go; fi

go test *.go
