#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

godep go build -o blog_app *.go
rm blog_app
