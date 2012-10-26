package main

import (
	"bytes"
	"encoding/json"
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
	logger.Noticef("buildArticlesCache(): %d admin articles, %d non admin", len(articlesCache.adminArticles), len(articlesCache.nonAdminArticles))

	js, sha1 := buildArticlesJson(adminArticles)
	articlesCache.adminArticlesJs, articlesCache.adminArticlesJsSha1 = js, sha1

	js, sha1 = buildArticlesJson(nonAdminArticles)
	articlesCache.nonAdminArticlesJs, articlesCache.nonAdminArticlesJsSha1 = js, sha1
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

func getArticlesJsData(isAdmin bool) ([]byte, string) {
	articlesCache.Lock()
	defer articlesCache.Unlock()

	if isAdmin {
		return articlesCache.adminArticlesJs, articlesCache.adminArticlesJsSha1
	}
	return articlesCache.nonAdminArticlesJs, articlesCache.nonAdminArticlesJsSha1
}

func getCachedArticles(isAdmin bool) []*Article {
	articlesCache.Lock()
	defer articlesCache.Unlock()

	logger.Noticef("getCachedArticles(): %d admin articles, %d non admin", len(articlesCache.adminArticles), len(articlesCache.nonAdminArticles))
	if isAdmin {
		return articlesCache.adminArticles
	}
	return articlesCache.nonAdminArticles
}
