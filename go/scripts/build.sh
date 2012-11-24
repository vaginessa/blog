export GOPATH=`pwd`/ext:$GOPATH
go build -o blog_app *.go
if [ "$?" -ne 0 ]; then echo "failed to build"; exit 1; fi
cp util.go extract_crashing_lines.go tools/importappengine
# only exists locally, not on the server
if [-e tools/importappengine ]
	cd tools/importappengine
	go build -o importappeng *.go
	if [ "$?" -ne 0 ]; then echo "failed to build"; exit 1; fi
	rm util.go extract_crashing_lines.go importappeng
fi

