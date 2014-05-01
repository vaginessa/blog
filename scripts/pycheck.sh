#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

pyflakes scripts/*py
