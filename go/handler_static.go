package main

import (
	"net/http"
	"path/filepath"
)

func getStaticDir() string {
	// when running locally
	d := filepath.Join("..", "appengine", "www", "static")
	if PathExists(d) {
		return d
	}
	// TODO: this will probably be different on the server
	logger.Errorf("getStaticDir(): '%s' dir doesn't exist", d)
	return ""
}

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

// url: /static/*
func handleStatic(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path[len("/static/"):]
	serveFileFromDir(w, r, getStaticDir(), file)
}

// url: /software*
func handleSoftware(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	if url == "/software" || url == "/software/" || url == "/software/index.html" {
		serveFileFromDir(w, r, getAppEngineTmplDir(), "software.html")
		return
	}
	if redirectIfNeeded(w, r) {
		return
	}
	file := r.URL.Path[len("/software/"):]
	serveFileFromDir(w, r, getSoftwareDir(), file)
}

// url: /favicon.ico
func handleFavicon(w http.ResponseWriter, r *http.Request) {
	serveFileFromDir(w, r, getStaticDir(), "favicon.ico")
}
