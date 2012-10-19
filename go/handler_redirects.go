package main

import (
	"net/http"
)

var redirects = map[string]string{
	"/kb/serialization-in-c#.html":        "/article/Serialization-in-C.html",
	"/articles/":                          "/articles/index.html",
	"/software/fofou":                     "/software/fofou/index.html",
	"/software/sumatra":                   "/software/sumatrapdf/free-pdf-reader.html",
	"/software/sumatrapdf":                "/software/sumatrapdf/free-pdf-reader.html",
	"/software/sumatrapdf/":               "/software/sumatrapdf/free-pdf-reader.html",
	"/software/sumatrapdf/index.html":     "/software/sumatrapdf/free-pdf-reader.html",
	"/software/sumatrapdf/download.html":  "/software/sumatrapdf/download-free-pdf-viewer.html",
	"/software/sumatrapdf/prerelase.html": "/software/sumatrapdf/prerelease.html",
	"/software/volante":                   "/software/volante/database.html",
	"/software/volante/":                  "/software/volante/database.html",
	"/software/volante/index.html":        "/software/volante/database.html",
	"/extremeoptimizations":               "/extremeoptimizations/index.html",
	"/extremeoptimizations/":              "/extremeoptimizations/index.html",
	"/atom.xml":                           "http://feeds.feedburner.com/KrzysztofKowalczykBlog",
	"/feed/rss2/atom.xml":                 "http://feeds.feedburner.com/KrzysztofKowalczykBlog",
	"/feed/rss2/":                         "http://feeds.feedburner.com/KrzysztofKowalczykBlog",
	"/feed/rss2":                          "http://feeds.feedburner.com/KrzysztofKowalczykBlog",
	"/feed/":                              "http://feeds.feedburner.com/KrzysztofKowalczykBlog",
	"/feed":                               "http://feeds.feedburner.com/KrzysztofKowalczykBlog",
	"/articles/cocoa-objectivec-reference.html":     "/articles/cocoa-reference.html",
	"/forum_sumatra/rss.php":                        "http://forums.fofou.org/sumatrapdf/rss",
	"/forum_sumatra":                                "http://forums.fofou.org/sumatrapdf",
	"/google6dba371684d43cd6.html":                  "/static/google6dba371684d43cd6.html",
	"/software/sumatrapdf/sumatra-shot-00.gif":      "http://kjkpub.s3.amazonaws.com/blog/sumatra/sumatra-shot-00.gif",
	"/software/sumatrapdf/sumatra-shot-01.gif":      "http://kjkpub.s3.amazonaws.com/blog/sumatra/sumatra-shot-01.gif",
	"/software/sumatrapdf/sumatra-shot-00-full.gif": "http://kjkpub.s3.amazonaws.com/blog/sumatra/sumatra-shot-00-full.gif",
	"/software/sumatrapdf/sumatra-shot-01-full.gif": "http://kjkpub.s3.amazonaws.com/blog/sumatra/sumatra-shot-01-full.gif",
	"/software/sumatrapdf/SumatraSplash.png":        "http://kjkpub.s3.amazonaws.com/blog/sumatra/SumatraSplash.png",
}

func redirectIfNeeded(w http.ResponseWriter, r *http.Request) bool {
	url := r.URL.Path
	//logger.Noticef("redirectIfNeeded(): '%s'", url)
	if redirUrl, ok := redirects[url]; ok {
		logger.Noticef("Redirecting '%s' => '%s'", url, redirUrl)
		http.Redirect(w, r, redirUrl, 302)
		return true
	}
	return false
}
