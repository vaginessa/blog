package main

import (
	"fmt"
	"flag"
	_ "html/template"
	"net/http"
	_ "net/url"
)

var (
	httpAddr   = flag.String("addr", ":8100", "HTTP server address")
)

func isTopLevelUrl(url string) bool {
	return 0 == len(url) || "/" == url
}

func serve404(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprint(w, `<html><body>Page Not Found!</body></html>`)
	http.NotFound(w, r)
}

// responds to /
func handleMain(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handleMain() %s\n", r.URL.Path)
	if !isTopLevelUrl(r.URL.Path) {
		serve404(w, r)
		return
	}
	fmt.Fprint(w, "This is /")
}

// responds to /blog
func handleBlogMain(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("handleBlogMain()\n")
}

func main() {
	flag.Parse()

	http.HandleFunc("/blog", handleBlogMain)
	http.HandleFunc("/", handleMain)

	fmt.Printf("Starting the blog on %s\n", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		fmt.Printf("http.ListendAndServer() failed with %s\n", err.Error())
	}
	fmt.Printf("Exited\n")

}
