package main

import (
	"atom"
	"fmt"
	"net/http"
	"time"
)

func handleAtom(w http.ResponseWriter, r *http.Request, excludeNotes bool) {
	articles := getCachedArticles(false)
	if excludeNotes {
		articles = filterArticlesByTag(articles, "note", false)
	}
	n := 25
	if n > len(articles) {
		n = len(articles)
	}

	latest := make([]*Article, n, n)
	size := len(articles)
	for i := 0; i < n; i++ {
		latest[i] = articles[size-1-i]
	}

	pubTime := time.Now()
	if len(articles) > 0 {
		pubTime = articles[0].PublishedOn
	}

	feed := &atom.Feed{
		Title:   "Krzysztof Kowalczyk blog",
		Link:    "http://blog.kowalczyk.info/atom.xml",
		PubDate: pubTime,
	}

	for _, a := range latest {

		ver := a.CurrVersion()
		msgHtml := articleBodyCache.GetHtml(ver.Sha1, ver.Format)

		id := fmt.Sprintf("tag:blog.kowalczyk.info,1999:%d", a.Id)
		e := &atom.Entry{
			Id:          id,
			Title:       a.Title,
			Link:        "http://blog.kowalczyk.info/" + a.Permalink(),
			ContentHtml: msgHtml,
			PubDate:     a.PublishedOn,
		}
		feed.AddEntry(e)
	}

	s, err := feed.GenXml()
	if err != nil {
		s = "Failed to generate XML feed"
	}

	w.Write([]byte(s))
}

// /atom-all.xml
func handleAtomAll(w http.ResponseWriter, r *http.Request) {
	handleAtom(w, r, false)
}

// /feedburner.xml
func handleFeedburnerAtom(w http.ResponseWriter, r *http.Request) {
	handleAtom(w, r, true)
}
