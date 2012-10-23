package main

import (
	"net/http"
)

// url: /
func handleMainPage(w http.ResponseWriter, r *http.Request) {
	logger.Noticef("handleMainPage(): %s", r.URL.Path)

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
		IsAdmin:     isAdmin,
		JqueryUrl:   "http://ajax.googleapis.com/ajax/libs/jquery/1.4.2/jquery.min.js",
		LogInOutUrl: getLogInOutUrl(r),
		Articles:    store.GetRecentArticles(10, isAdmin),
	}

	ExecTemplate(w, tmplMainPage, model)
}
