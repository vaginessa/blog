package main

import (
	"net/http"
)

type Year struct {
	Name     string
	Articles []MonthArticle
}

type MonthArticle struct {
	*Article
	DisplayMonth string
}

type ArticlesIndexModel struct {
	IsAdmin       bool
	AnalyticsCode string
	JqueryUrl     string
	LogInOutUrl   string
	ArticlesJsUrl string
	PostsCount    int
	Tag           string
	Years         []Year
}

func (a *MonthArticle) DisplayTitle() string {
	if a.Title != "" {
		return a.Title
	}
	return "no title"
}

func NewYear(name string) *Year {
	return &Year{Name: name, Articles: make([]MonthArticle, 0)}
}

func buildYearsFromArticles(articles []*Article) []Year {
	res := make([]Year, 0)
	var currYear *Year
	var currMonthName string
	n := len(articles)
	for i := n - 1; i >= 0; i-- {
		a := articles[i]
		yearName := a.PublishedOn().Format("2006")
		if currYear == nil || currYear.Name != yearName {
			if currYear != nil {
				res = append(res, *currYear)
			}
			currYear = NewYear(yearName)
			currMonthName = ""
		}
		ma := MonthArticle{Article: a}
		monthName := a.PublishedOn().Format("01")
		if monthName != currMonthName {
			ma.DisplayMonth = a.PublishedOn().Format("January 2")
		} else {
			ma.DisplayMonth = a.PublishedOn().Format("2")
		}
		currMonthName = monthName
		currYear.Articles = append(currYear.Articles, ma)
	}
	if currYear != nil {
		res = append(res, *currYear)
	}
	return res
}

func filterArticlesByTag(articles []*Article, tag string) []*Article {
	res := make([]*Article, 0)
	for _, a := range articles {
		for _, t := range a.Tags {
			if tag == t {
				res = append(res, a)
				break
			}
		}
	}
	return res
}

func showArchivePage(w http.ResponseWriter, r *http.Request, tag string) {
	isAdmin := IsAdmin(r)
	// must be called first as it builds the cache if needed
	articlesJsUrl := getArticlesJsUrl(isAdmin)
	articles := getCachedArticles(isAdmin)

	if tag != "" {
		articles = filterArticlesByTag(articles, tag)
	}

	model := ArticlesIndexModel{
		IsAdmin:       isAdmin,
		JqueryUrl:     jQueryUrl(),
		LogInOutUrl:   getLogInOutUrl(r),
		ArticlesJsUrl: articlesJsUrl,
		PostsCount:    len(articles),
		Years:         buildYearsFromArticles(articles),
		Tag:           tag,
	}

	ExecTemplate(w, tmplArchive, model)
}

// url: /tag/${tag}
func handleTag(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Path[len("/tag/"):]
	showArchivePage(w, r, tag)
}

// url: /archives.html
func handleArchives(w http.ResponseWriter, r *http.Request) {
	showArchivePage(w, r, "")
}
