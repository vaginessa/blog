package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type DisplayArticle struct {
	*Article
	HtmlBody template.HTML
}

// TODO: this is simplistic but works for me, http://net.tutsplus.com/tutorials/other/8-regular-expressions-you-should-know/
// has more elaborate regex for extracting urls
var urlRx = regexp.MustCompile(`https?://[[:^space:]]+`)
var notUrlEndChars = []byte(".),")

func notUrlEndChar(c byte) bool {
	return -1 != bytes.IndexByte(notUrlEndChars, c)
}

var disableUrlization = false

func strToHtml(s string) string {
	matches := urlRx.FindAllStringIndex(s, -1)
	if nil == matches || disableUrlization {
		s = template.HTMLEscapeString(s)
		s = strings.Replace(s, "\n", "<br>", -1)
		return s
	}

	urlMap := make(map[string]string)
	ns := ""
	prevEnd := 0
	for n, match := range matches {
		start, end := match[0], match[1]
		for end > start && notUrlEndChar(s[end-1]) {
			end -= 1
		}
		url := s[start:end]
		ns += s[prevEnd:start]

		// placeHolder is meant to be an unlikely string that doesn't exist in
		// the message, so that we can replace the string with it and then
		// revert the replacement. A more robust approach would be to remember
		// offsets
		placeHolder, ok := urlMap[url]
		if !ok {
			placeHolder = fmt.Sprintf("a;dfsl;a__lkasjdfh1234098;lajksdf_%d", n)
			urlMap[url] = placeHolder
		}
		ns += placeHolder
		prevEnd = end
	}

	ns = template.HTMLEscapeString(ns)
	for url, placeHolder := range urlMap {
		url = fmt.Sprintf(`<a href="%s" rel="nofollow">%s</a>`, url, url)
		ns = strings.Replace(ns, placeHolder, url, -1)
	}
	ns = strings.Replace(ns, "\n", "<br>", -1)
	return ns
}

func msgToHtml(msg string, format int) string {
	if format == FormatHtml {
		return msg
	}
	if format == FormatTextile {
		// TODO: convert textile to html
		return msg
	}
	if format == FormatMarkdown {
		// TODO: convert markdown to html
		return msg
	}
	if format == FormatText {
		return strToHtml(msg)
	}
	panic("unknown format")
	return ""
}

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

	displayArticle := &DisplayArticle{Article: article}

	ver := article.Versions[len(article.Versions)-1]
	msgFilePath := store.MessageFilePath(ver.Sha1)
	msg, err := ioutil.ReadFile(msgFilePath)
	msgHtml := ""
	if err != nil {
		msgHtml = fmt.Sprintf("Error: failed to fetch a message with sha1 %x, file: %s", ver.Sha1[:], msgFilePath)
	} else {
		msgHtml = msgToHtml(string(msg), ver.Format)
	}

	displayArticle.HtmlBody = template.HTML(msgHtml)

	model := struct {
		IsAdmin        bool
		AnalyticsCode  string
		JqueryUrl      string
		PageTitle      string
		Article        *DisplayArticle
		NextArticle    *DisplayArticle
		PrevArticle    *DisplayArticle
		LogInOutUrl    string
		ArticlesJsUrl  string
		PrettifyJsUrl  string
		PrettifyCssUrl string
		TagsDisplay    string
		ArticleNo      int
		ArticlesCount  int
	}{
		IsAdmin:       isAdmin,
		JqueryUrl:     jQueryUrl(),
		LogInOutUrl:   getLogInOutUrl(r),
		Article:       displayArticle,
		PageTitle:     article.Title,
		ArticlesCount: store.ArticlesCount(),
		ArticlesJsUrl: getArticlesJsUrl(isAdmin),
	}

	ExecTemplate(w, tmplArticle, model)
}
