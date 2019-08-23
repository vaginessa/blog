#!/bin/bash
set -u -e -o pipefail

wget https://github.com/netlify/netlifyctl/releases/download/v0.4.0/netlifyctl-linux-amd64-0.4.0.tar.gz
tar -xvf netlifyctl-linux-amd64-0.4.0.tar.gz

go build -o blog
./blog -deploy
./netlifyctl -A $NETLIFY_TOKEN deploy || true
cat netlifyctl-debug.log || true
