#!/bin/bash
cp util.go extract_crashing_lines.go tools/importappengine
cd tools/importappengine
go build -o importappeng *.go
if [ "$?" -ne 0 ]; then echo "failed to build"; exit 1; fi
rm util.go extract_crashing_lines.go # we used you so now we discard you
cd ../..
./tools/importappengine/importappeng
rm tools/importappengine/importappeng
