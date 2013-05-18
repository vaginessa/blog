cd go
export GOPATH=`pwd`/ext:$GOPATH
#go run *.go
go build -o blog_app *.go
./blog_app
