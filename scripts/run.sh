cd go
export GOPATH=`pwd`/ext:$GOPATH
#go run *.go
go build -o blog_app *.go || exit 1
./blog_app
