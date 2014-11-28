package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/kjk/u"
)

var articlesCache ArticlesCache

type ArticlesCache struct {
	sync.Mutex
	articlesCacheId        int
	nonAdminArticles       []*Article
	nonAdminArticlesJs     []byte
	nonAdminArticlesJsSha1 string
}

func appendJsonMarshalled(buf *bytes.Buffer, val interface{}) {
	if data, err := json.Marshal(val); err != nil {
		logger.Errorf("json.Marshal() of %v failed with %s", val, err)
	} else {
		buf.Write(data)
	}
}

// TODO: I only use it for tag cloud, could just send info about tags directly
func buildArticlesJson(articles []*Article) ([]byte, string) {
	var buf bytes.Buffer
	buf.WriteString("var __articles_json = ")
	n := len(articles)
	vals := make([]interface{}, n, n)
	n = 0
	for i := len(articles) - 1; i >= 0; i-- {
		a := articles[i]
		/*
			val := make([]interface{}, 6, 6)
			val[0] = a.PublishedOn.Format("2006-01-02")
			val[1] = a.Permalink()
			val[2] = a.Title
			val[3] = a.Tags
			val[4] = !a.IsPrivate
			val[5] = a.IsDeleted
		*/
		val := make([]interface{}, 1, 1)
		val[0] = a.Tags
		vals[n] = val
		n += 1
	}
	appendJsonMarshalled(&buf, vals)
	buf.WriteString("; articlesJsonLoaded(__articles_json);")
	jsData := buf.Bytes()
	sha1 := u.Sha1HexOfBytes(jsData)
	//logger.Noticef("buildArticlesJson(): len(jsData)=%d, sha1=%s", len(jsData), sha1)
	return jsData, sha1
}

// must be called with a articlesCache locked
func buildArticlesCache(articlesCacheId int, articles []*Article) {
	nonAdminArticles := make([]*Article, 0)
	for _, a := range articles {
		nonAdminArticles = append(nonAdminArticles, a)
	}
	articlesCache.articlesCacheId = articlesCacheId
	articlesCache.nonAdminArticles = nonAdminArticles
	js, sha1 := buildArticlesJson(nonAdminArticles)
	articlesCache.nonAdminArticlesJs, articlesCache.nonAdminArticlesJsSha1 = js, sha1
}

func rebuildArticlesCacheIfNeededUnlocked() {
	articlesCacheId, articles := store.GetArticles(articlesCache.articlesCacheId)
	if articlesCacheId != articlesCache.articlesCacheId {
		buildArticlesCache(articlesCacheId, articles)
	}
}

func getArticlesJsUrl(isAdmin bool) string {
	articlesCache.Lock()
	defer articlesCache.Unlock()
	rebuildArticlesCacheIfNeededUnlocked()
	sha1 := articlesCache.nonAdminArticlesJsSha1
	return "/djs/articles-" + sha1 + ".js"
}

func getArticlesJsData(isAdmin bool) ([]byte, string) {
	articlesCache.Lock()
	defer articlesCache.Unlock()

	rebuildArticlesCacheIfNeededUnlocked()
	return articlesCache.nonAdminArticlesJs, articlesCache.nonAdminArticlesJsSha1
}

func getCachedArticles(isAdmin bool) []*Article {
	articlesCache.Lock()
	defer articlesCache.Unlock()

	rebuildArticlesCacheIfNeededUnlocked()
	return articlesCache.nonAdminArticles
}

func getCachedArticlesById(articleId int, isAdmin bool) (*Article, *Article, *Article, int) {
	articles := getCachedArticles(isAdmin)
	var prev, next *Article
	for i, curr := range articles {
		if curr.Id == articleId {
			if i != len(articles)-1 {
				next = articles[i+1]
			}
			return prev, curr, next, i
		}
		prev = curr
	}
	return nil, nil, nil, 0
}

func GetArticleHtml(bodyId string, format int) string {
	msgHtml := store.GetArticleHtml(bodyId)
	if msgHtml == "" {
		msgHtml = fmt.Sprintf("Error: failed to fetch a message with bodyId %q", bodyId)
	}
	return msgHtml
}
