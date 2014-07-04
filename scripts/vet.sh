#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

go vet -printfuncs=httpErrorf:1,panicif:1,Noticef,Errorf *.go
