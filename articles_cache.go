package main

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/kjk/u"
)

var articlesCache ArticlesCache

type ArticlesCache struct {
	sync.Mutex
	articles       []*Article
	articlesJs     []byte
	articlesJsSha1 string
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

func buildArticlesCache() {
	u.PanicIf(articlesCache.articles != nil)
	articles := store.GetArticles()
	articlesCache.articles = articles
	articlesCache.articlesJs, articlesCache.articlesJsSha1 = buildArticlesJson(articles)
}

func getArticlesJsUrl() string {
	sha1 := articlesCache.articlesJsSha1
	return "/djs/articles-" + sha1 + ".js"
}

func getArticlesJsData() ([]byte, string) {
	return articlesCache.articlesJs, articlesCache.articlesJsSha1
}

func getCachedArticles() []*Article {
	return articlesCache.articles
}

func getCachedArticlesById(articleId int) (*Article, *Article, *Article, int) {
	articles := getCachedArticles()
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
