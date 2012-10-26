package main

import (
	"net/http"
	"strconv"
	"strings"
)

// url is in the form: $sha1.js
func handleArticlesJs(w http.ResponseWriter, r *http.Request, url string) {
	sha1 := url[:len(url)-len(".js")]
	if len(sha1) != 40 {
		logger.Errorf("handleArticlesJs(): invalid sha1='%s', url='%s", sha1, url)
		panic("invalid sha1")
	}

	jsData, expectedSha1 := getArticlesJsData(IsAdmin(r))
	if sha1 != expectedSha1 {
		logger.Errorf("handleArticlesJs(): invalid value of sha1='%s', expected='%s'", sha1, expectedSha1)
		panic("invalid value of sha1")
	}

	w.Header().Set("Content-Type", "text/javascript")
	// cache non-admin version by setting max age 1 year into the future
	// http://betterexplained.com/articles/how-to-optimize-your-site-with-http-caching/
	if !IsAdmin(r) {
		w.Header().Set("Cache-Control", "max-age=31536000, public")
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(jsData)))
	w.Write(jsData)
}

// url: /djs/$url
func handleDjs(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path[len("/djs/"):]
	if strings.HasPrefix(url, "articles-") {
		handleArticlesJs(w, r, url[len("articles-"):])
		return
	}
	serve404(w, r)
}

// url: /
func handleMainPage(w http.ResponseWriter, r *http.Request) {
	if !isTopLevelUrl(r.URL.Path) {
		serve404(w, r)
		return
	}

	isAdmin := IsAdmin(r)
	model := struct {
		IsAdmin       bool
		AnalyticsCode string
		JqueryUrl     string
		Articles      []*Article
		LogInOutUrl   string
		ArticlesJsUrl string
	}{
		IsAdmin:       isAdmin,
		AnalyticsCode: *config.AnalyticsCode,
		JqueryUrl:     jQueryUrl(),
		LogInOutUrl:   getLogInOutUrl(r),
		Articles:      store.GetRecentArticles(10, isAdmin),
		ArticlesJsUrl: getArticlesJsUrl(isAdmin),
	}

	ExecTemplate(w, tmplMainPage, model)
}
