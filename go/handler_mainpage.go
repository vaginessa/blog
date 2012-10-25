package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

type ArticlesCache struct {
	sync.Mutex
	articlesCacheId        int
	adminArticles          []*Article
	adminArticlesJs        []byte
	adminArticlesJsSha1    string
	nonAdminArticles       []*Article
	nonAdminArticlesJs     []byte
	nonAdminArticlesJsSha1 string
}

var articlesCache ArticlesCache

func appendJsonMarshalled(buf *bytes.Buffer, val interface{}) {
	if data, err := json.Marshal(val); err != nil {
		logger.Errorf("json.Marshal() of %v failed with %s", val, err.Error())
	} else {
		buf.Write(data)
	}
}

func buildArticlesJson(articles []*Article) ([]byte, string) {
	var buf bytes.Buffer
	buf.WriteString("var __articles_json = ")
	n := len(articles)
	vals := make([]interface{}, n, n)
	n = 0
	for i := len(articles) - 1; i >= 0; i-- {
		a := articles[i]
		val := make([]interface{}, 6, 6)
		val[0] = a.PublishedOn().Format("2006-01-02")
		val[1] = a.Permalink()
		val[2] = a.Title
		val[3] = a.Tags
		val[4] = !a.IsPrivate
		val[5] = a.IsDeleted
		vals[n] = val
		n += 1
	}
	appendJsonMarshalled(&buf, vals)
	buf.WriteString("; articlesJsonLoaded(__articles_json);")
	jsData := buf.Bytes()
	sha1 := Sha1StringOfBytes(jsData)
	logger.Noticef("buildArticlesJson(): len(jsData)=%d, sha1=%s", len(jsData), sha1)
	return buf.Bytes(), Sha1StringOfBytes(buf.Bytes())
}

// must be called with a articlesCache locked
func buildArticlesCache(articlesCacheId int, articles []Article) {
	n := len(articles)
	adminArticles := make([]*Article, n, n)
	nonAdminArticles := make([]*Article, 0)
	for i, a := range articles {
		article := &articles[i]
		adminArticles[i] = article
		if !a.IsDeleted && !a.IsPrivate {
			nonAdminArticles = append(nonAdminArticles, article)
		}
	}

	articlesCache.articlesCacheId = articlesCacheId
	articlesCache.adminArticles = adminArticles
	articlesCache.nonAdminArticles = nonAdminArticles

	js, sha1 := buildArticlesJson(adminArticles)
	articlesCache.adminArticlesJs, articlesCache.adminArticlesJsSha1 = js, sha1

	js, sha1 = buildArticlesJson(nonAdminArticles)
	articlesCache.nonAdminArticlesJs, articlesCache.nonAdminArticlesJsSha1 = js, sha1
}

// url is in the form: $sha1.js
func handleArticlesJs(w http.ResponseWriter, r *http.Request, url string) {
	articlesCache.Lock()
	defer articlesCache.Unlock()

	sha1 := url[:len(url)-len(".js")]
	if len(sha1) != 40 {
		logger.Errorf("handleArticlesJs(): invalid sha1='%s', url='%s", sha1, url)
		panic("invalid sha1")
	}
	var jsData []byte
	var expectedSha1 string
	if IsAdmin(r) {
		jsData = articlesCache.adminArticlesJs
		expectedSha1 = articlesCache.adminArticlesJsSha1
	} else {
		jsData = articlesCache.nonAdminArticlesJs
		expectedSha1 = articlesCache.nonAdminArticlesJsSha1
	}
	if sha1 != expectedSha1 {
		logger.Errorf("handleArticlesJs(): invalid value of sha1='%s', expected='%s'", sha1, expectedSha1)
		panic("invalid value of sha1")
	}

	w.Header().Set("Content-Type", "text/javascript")

	// TODO: set expiration in the future
	/*
	   # must over-ride Cache-Control (is 'no-cache' by default)
	   self.response.headers['Cache-Control'] = 'public, max-age=31536000'
	   now = datetime.datetime.now()
	   expires_date_txt = httpdate(now + datetime.timedelta(days=365))
	   self.response.headers.add_header("Expires", expires_date_txt)
	*/

	w.Write(jsData)
}

// url: /djs/$url
func handleDjs(w http.ResponseWriter, r *http.Request) {
	logger.Noticef("handleDjs(): %s", r.URL.Path)
	url := r.URL.Path[len("/djs/"):]
	if strings.HasPrefix(url, "articles-") {
		handleArticlesJs(w, r, url[len("articles-"):])
		return
	}
	serve404(w, r)
}

func getArticlesJsUrl(isAdmin bool) string {
	articlesCache.Lock()
	defer articlesCache.Unlock()

	articlesCacheId, articles := store.GetArticles(articlesCache.articlesCacheId)
	if articlesCacheId != articlesCache.articlesCacheId {
		logger.Notice("getArticlesJsUrl(): rebuilding articlesCache")
		buildArticlesCache(articlesCacheId, articles)
	} else {
		logger.Notice("getArticlesJsUrl(): articlesCache unchanged")
	}
	var sha1 string
	if isAdmin {
		sha1 = articlesCache.adminArticlesJsSha1
	} else {
		sha1 = articlesCache.nonAdminArticlesJsSha1
	}
	return "/djs/articles-" + sha1 + ".js"
}

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
		IsAdmin:       isAdmin,
		JqueryUrl:     jQueryUrl(),
		LogInOutUrl:   getLogInOutUrl(r),
		Articles:      store.GetRecentArticles(10, isAdmin),
		ArticlesJsUrl: getArticlesJsUrl(isAdmin),
	}

	ExecTemplate(w, tmplMainPage, model)
}
