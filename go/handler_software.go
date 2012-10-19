package main

import (
	_ "code.google.com/p/gorilla/mux"
	"net/http"
	"path/filepath"
	"strings"
)

func getSoftwareDir() string {
	// when running locally
	d := filepath.Join("..", "appengine", "www", "software")
	if PathExists(d) {
		return d
	}
	// TODO: this will probably be different on the server
	logger.Errorf("getSoftwareDir(): '%s' dir doesn't exist", d)
	return ""
}

func getAppEngineTmplDir() string {
	// when running locally
	d := filepath.Join("..", "appengine", "tmpl")
	if PathExists(d) {
		return d
	}
	// TODO: this will probably be different on the server
	logger.Errorf("getAppEngineTmplDir(): '%s' dir doesn't exist", d)
	return ""
}

func serveFileFromDir(w http.ResponseWriter, r *http.Request, dir, fileName string) {
	filePath := filepath.Join(dir, fileName)
	logger.Noticef("serveFileFromDir(): '%s'", filePath)
	if !PathExists(filePath) {
		logger.Noticef("serveFileFromDir() file '%s' doesn't exist, referer: '%s'", fileName, getReferer(r))
	}
	http.ServeFile(w, r, filePath)
}

// url: /software, /software/, /software/index.html
func handleSoftwareIndex(w http.ResponseWriter, r *http.Request) {
	if redirectIfNeeded(w, r) {
		return
	}
	serveFileFromDir(w, r, getAppEngineTmplDir(), "software.html")
}

// url: /software/{program}[/{rest}]
func handleSoftware(w http.ResponseWriter, r *http.Request) {
	if redirectIfNeeded(w, r) {
		return
	}
	file := r.URL.Path[len("/software/"):]
	if strings.HasSuffix(file, "/") {
		file += "index.html"
	}
	serveFileFromDir(w, r, getSoftwareDir(), file)
}
