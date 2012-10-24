package main

import (
	"net/http"
	"strings"
)

// url: /article/*
func handleArticle(w http.ResponseWriter, r *http.Request) {
	logger.Noticef("handleArticle(): %s", r.URL.Path)
	url := r.URL.Path
	isAdmin := IsAdmin(r)

	// we expect /article/$shortId/$url
	parts := strings.SplitN(url[len("/article/"):], "/", 2)
	if len(parts) != 2 {
		logger.Noticef("parts: %v", parts)
		serve404(w, r)
		return
	}

	articleId := UnshortenId(parts[0])
	logger.Noticef("article id: %d", articleId)
	article := store.GetArticleById(articleId)
	if nil == article {
		serve404(w, r)
		return
	}

	model := struct {
		IsAdmin        bool
		AnalyticsCode  string
		JqueryUrl      string
		PageTitle      string
		Article        *Article
		NextArticle    *Article
		PrevArticle    *Article
		LogInOutUrl    string
		ArticlesJsUrl  string
		PrettifyJsUrl  string
		PrettifyCssUrl string
		TagsDisplay    string
		ArticleNo      int
		ArticlesCount  int
	}{
		IsAdmin:     isAdmin,
		JqueryUrl:   "http://ajax.googleapis.com/ajax/libs/jquery/1.4.2/jquery.min.js",
		LogInOutUrl: getLogInOutUrl(r),
		Article:     article,
		PageTitle:   article.Title,
	}

	ExecTemplate(w, tmplArticle, model)
}
