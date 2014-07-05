package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"sync"

	"github.com/kjk/textiler"
	"github.com/kjk/u"
	"github.com/russross/blackfriday"
)

var articleBodyCache ArticleBodyCache
var articlesCache ArticlesCache

type ArticlesCache struct {
	sync.Mutex
	articlesCacheId        int
	adminArticles          []*Article2
	adminArticlesJs        []byte
	adminArticlesJsSha1    string
	nonAdminArticles       []*Article2
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
func buildArticlesJson(articles []*Article2) ([]byte, string) {
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
	sha1 := u.Sha1StringOfBytes(jsData)
	//logger.Noticef("buildArticlesJson(): len(jsData)=%d, sha1=%s", len(jsData), sha1)
	return buf.Bytes(), u.Sha1StringOfBytes(buf.Bytes())
}

// must be called with a articlesCache locked
func buildArticlesCache(articlesCacheId int, articles []*Article2) {
	n := len(articles)
	adminArticles := make([]*Article2, n, n)
	nonAdminArticles := make([]*Article2, 0)
	for i, a := range articles {
		adminArticles[i] = a
		if !a.IsDeleted && !a.IsPrivate {
			nonAdminArticles = append(nonAdminArticles, a)
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

	rebuildArticlesCacheIfNeededUnlocked()
	if isAdmin {
		return articlesCache.adminArticlesJs, articlesCache.adminArticlesJsSha1
	}
	return articlesCache.nonAdminArticlesJs, articlesCache.nonAdminArticlesJsSha1
}

func getCachedArticles(isAdmin bool) []*Article2 {
	articlesCache.Lock()
	defer articlesCache.Unlock()

	rebuildArticlesCacheIfNeededUnlocked()
	if isAdmin {
		return articlesCache.adminArticles
	}
	return articlesCache.nonAdminArticles
}

func getCachedArticlesById(articleId int, isAdmin bool) (*Article2, *Article2, *Article2, int) {
	articles := getCachedArticles(isAdmin)
	var prev, next *Article2
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

type ArticleBodyCacheEntry struct {
	bodyId  string
	msgHtml string
}

type ArticleBodyCache struct {
	sync.Mutex
	entries      [64]ArticleBodyCacheEntry
	entriesCount int
	curr         int
}

func (c *ArticleBodyCache) GetHtml(bodyId string, format int) string {
	c.Lock()
	defer c.Unlock()

	for i := 0; i < c.entriesCount; i++ {
		if c.entries[i].bodyId == bodyId {
			return c.entries[i].msgHtml
		}
	}

	msg, err := store.GetTextBody(bodyId)
	var msgHtml string
	if err != nil {
		msgHtml = fmt.Sprintf("Error: failed to fetch a message with bodyId %q", bodyId)
	} else {
		msgHtml = msgToHtml(msg, format)
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

	entry.bodyId = bodyId
	entry.msgHtml = msgHtml
	return msgHtml
}

func (c *ArticleBodyCache) Clear() {
	c.Lock()
	defer c.Unlock()
	c.entriesCount = 0
	c.curr = 0
}

func clearArticlesCache() {
	articlesCache.articlesCacheId = 0
	articleBodyCache.Clear()
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
		return "<p>" + s + "</p>"
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
	ns += s[prevEnd:len(s)]

	ns = template.HTMLEscapeString(ns)
	for url, placeHolder := range urlMap {
		url = fmt.Sprintf(`<a href="%s" rel="nofollow">%s</a>`, url, url)
		ns = strings.Replace(ns, placeHolder, url, -1)
	}
	ns = strings.Replace(ns, "\n", "<br>", -1)
	return "<p>" + ns + "</p>"
}

func textile(s []byte) string {
	s, replacements := txt_with_code_parts(s)
	res := textiler.ToHtml(s, false, false)
	for kStr, v := range replacements {
		k := []byte(kStr)
		res = bytes.Replace(res, k, v, -1)
	}
	return string(res)
}

func markdown(s []byte) string {
	//fmt.Printf("msgToHtml(): markdown\n")
	s, replacements := txt_with_code_parts(s)
	renderer := blackfriday.HtmlRenderer(0, "", "")
	res := blackfriday.Markdown(s, renderer, 0)
	for kStr, v := range replacements {
		k := []byte(kStr)
		res = bytes.Replace(res, k, v, -1)
	}
	return string(res)
}

func msgToHtml(msg []byte, format int) string {
	switch format {
	case FormatHtml:
		//fmt.Printf("msgToHtml(): html\n")
		return string(msg)
	case FormatTextile:
		return textile(msg)
	case FormatMarkdown:
		return markdown(msg)
	case FormatText:
		//fmt.Printf("msgToHtml(): text\n")
		return strToHtml(string(msg))
	}
	panic("unknown format")
}
