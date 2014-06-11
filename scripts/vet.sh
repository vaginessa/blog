#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

if [ -e go ]; then cd go; fi
go tool vet  -printfuncs=httpErrorf:1,panicif:1,Noticef,Errorf *.go
