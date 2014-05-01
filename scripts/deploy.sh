#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

cd go
fab deploy
