package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type DisplayArticle struct {
	*Article
	HtmlBody template.HTML
}

func (a *DisplayArticle) PublishedOnShort() string {
	return a.PublishedOn().Format("Jan 2 2006")
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

type ArticleBodyCacheEntry struct {
	sha1    [20]byte
	msgHtml string
}

type ArticleBodyCache struct {
	sync.Mutex
	entries      [64]ArticleBodyCacheEntry
	entriesCount int
	curr         int
}

func (c *ArticleBodyCache) GetHtml(sha1 [20]byte, format int) string {
	c.Lock()
	defer c.Unlock()

	for i := 0; i < c.entriesCount; i++ {
		if c.entries[i].sha1 == sha1 {
			return c.entries[i].msgHtml
		}
	}

	msgFilePath := store.MessageFilePath(sha1)
	msg, err := ioutil.ReadFile(msgFilePath)
	var msgHtml string
	if err != nil {
		msgHtml = fmt.Sprintf("Error: failed to fetch a message with sha1 %x, file: %s", sha1[:], msgFilePath)
	} else {
		msgHtml = msgToHtml(string(msg), format)
	}

	var entry *ArticleBodyCacheEntry
	if c.entriesCount < len(c.entries) {
		entry = &c.entries[c.entriesCount]
		c.entriesCount += 1
	} else {
		entry = &c.entries[c.curr]
		c.curr += 1
		c.curr = c.curr % len(c.entries)
	}

	entry.sha1 = sha1
	entry.msgHtml = msgHtml
	return msgHtml
}

var articleBodyCache ArticleBodyCache

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
	prev, article, next, pos := getCachedArticlesById(articleId, isAdmin)
	if nil == article {
		serve404(w, r)
		return
	}

	logger.Noticef("%v, %v, %v, %d", prev, article, next, pos)

	displayArticle := &DisplayArticle{Article: article}

	ver := article.CurrVersion()
	msgHtml := articleBodyCache.GetHtml(ver.Sha1, ver.Format)
	displayArticle.HtmlBody = template.HTML(msgHtml)

	model := struct {
		IsAdmin        bool
		AnalyticsCode  string
		JqueryUrl      string
		PageTitle      string
		Article        *DisplayArticle
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
		IsAdmin:       isAdmin,
		AnalyticsCode: *config.AnalyticsCode,
		JqueryUrl:     jQueryUrl(),
		LogInOutUrl:   getLogInOutUrl(r),
		Article:       displayArticle,
		NextArticle:   next,
		PrevArticle:   prev,
		PageTitle:     article.Title,
		ArticlesCount: store.ArticlesCount(),
		ArticleNo:     pos + 1,
		ArticlesJsUrl: getArticlesJsUrl(isAdmin),
	}

	ExecTemplate(w, tmplArticle, model)
}
