# on the server the hierarchy is different
if [ -e go ]; then cd go; fi

export GOPATH=`pwd`/ext:$GOPATH
go test *.go
