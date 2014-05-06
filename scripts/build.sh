#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

# on the server the hierarchy is different
if [ -e go ]; then cd go; fi

go build -o blog_app *.go
# only exists locally, not on the server
if [ -e tools/importappengine ]; then
	cp util.go extract_crashing_lines.go tools/importappengine
	cd tools/importappengine
	go build -o importappeng *.go
	rm util.go extract_crashing_lines.go importappeng
fi
