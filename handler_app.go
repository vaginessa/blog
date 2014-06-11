package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	NOTE_TAG = "note"
)

func getTrimmedFormValue(r *http.Request, name string) string {
	return strings.TrimSpace(r.FormValue(name))
}

// /app/preview
func handleAppPreview(w http.ResponseWriter, r *http.Request) {
	format := getTrimmedFormValue(r, "format")
	formatInt := FormatNameToId(format)
	// TODO: what to do on error?
	msg := getTrimmedFormValue(r, "note")
	s := msgToHtml([]byte(msg), formatInt)
	//w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(s))
}

func checkboxToBool(checkboxVal string) bool {
	return "on" == checkboxVal
}

func tagsFromString(s string) []string {
	tags := strings.Split(s, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}
	return tags
}

// POST /app/edit
func createNewOrUpdatePost(w http.ResponseWriter, r *http.Request, article *Article) {
	format := FormatNameToId(getTrimmedFormValue(r, "format"))
	if !validFormat(format) {
		httpErrorf(w, "invalid format")
		return
	}
	title := getTrimmedFormValue(r, "title")
	if title == "" {
		httpErrorf(w, "empty title not valid")
		return
	}
	body := getTrimmedFormValue(r, "note")
	if len(body) < 10 {
		httpErrorf(w, "body too small")
		return
	}
	isPrivate := checkboxToBool(getTrimmedFormValue(r, "private_checkbox"))
	tags := tagsFromString(getTrimmedFormValue(r, "tags"))

	text, err := store.CreateNewText(format, body)
	if err != nil {
		logger.Errorf("createNewOrUpdatePost(): store.CreateNewText() failed with %s", err)
		httpErrorf(w, "error creating text")
		return
	}
	if article == nil {
		article = &Article{
			Id:          0,
			PublishedOn: time.Now(),
			Versions:    make([]*Text, 0),
		}
	}
	article.Versions = append(article.Versions, text)
	updatePublishedOn := checkboxToBool(getTrimmedFormValue(r, "update_published_on"))
	if updatePublishedOn {
		article.PublishedOn = time.Now()
	}
	article.Title = title
	article.IsPrivate = isPrivate
	article.IsDeleted = false
	article.Tags = tags
	if article, err = store.CreateOrUpdateArticle(article); err != nil {
		logger.Errorf("createNewOrUpdatePost(): store.CreateNewArticle() failed with %s", err)
		httpErrorf(w, "error creating article")
		return
	}
	clearArticlesCache()
	url := "/" + article.Permalink()
	http.Redirect(w, r, url, 301)
}

func GetArticleVersionBody(sha1 []byte) (string, error) {
	msgFilePath := store.MessageFilePath(sha1)
	msg, err := ioutil.ReadFile(msgFilePath)
	if err != nil {
		return "", err
	}
	return string(msg), nil
}

// url: /app/edit
func handleAppEdit(w http.ResponseWriter, r *http.Request) {
	if !IsAdmin(r) {
		http404(w, r)
		return
	}

	tags := make([]string, 0)
	if getTrimmedFormValue(r, "note") == "yes" {
		tags = append(tags, NOTE_TAG)
	}

	var article *Article
	articleIdStr := getTrimmedFormValue(r, "article_id")
	if articleId, err := strconv.Atoi(articleIdStr); err == nil {
		article = store.GetArticleById(articleId)
	}

	if r.Method == "POST" {
		createNewOrUpdatePost(w, r, article)
		return
	}

	model := struct {
		PrettifyCssUrl         string
		PrettifyJsUrl          string
		JqueryUrl              string
		FormatTextileChecked   string
		FormatMarkdownChecked  string
		FormatHtmlChecked      string
		FormatTextChecked      string
		PrivateCheckboxChecked string
		SubmitButtonText       string
		ArticleId              int
		ArticleTitle           string
		ArticleBody            template.HTML
		Tags                   string
	}{
		JqueryUrl:      jQueryUrl(),
		PrettifyJsUrl:  prettifyJsUrl(),
		PrettifyCssUrl: prettifyCssUrl(),
	}

	if article == nil {
		model.FormatMarkdownChecked = "selected"
		model.PrivateCheckboxChecked = "checked"
		model.SubmitButtonText = "Post"
		model.Tags = strings.Join(tags, ",")
	} else {
		model.ArticleId = article.Id
		model.ArticleTitle = article.Title
		ver := article.CurrVersion()
		if body, err := GetArticleVersionBody(ver.Sha1[:]); err != nil {
			panic("GetArticleVersionBody() failed")
		} else {
			model.ArticleBody = template.HTML(body)
		}
		model.SubmitButtonText = "Update post"
		model.Tags = strings.Join(article.Tags, ",")
		if article.IsPrivate {
			model.PrivateCheckboxChecked = "checked"
		}
		format := article.CurrVersion().Format
		checked := &model.FormatTextChecked
		if format == FormatHtml {
			checked = &model.FormatHtmlChecked
		} else if format == FormatTextile {
			checked = &model.FormatTextileChecked
		} else if format == FormatMarkdown {
			checked = &model.FormatMarkdownChecked
		} else if format != FormatText {
			panic("invalid format")
		}
		*checked = "selected"
	}

	ExecTemplate(w, tmplEdit, model)
}

func findArticleMustBeAdmin(w http.ResponseWriter, r *http.Request) *Article {
	if !IsAdmin(r) {
		http404(w, r)
		return nil
	}

	var article *Article
	idStr := getTrimmedFormValue(r, "article_id")
	if articleId, err := strconv.Atoi(idStr); err == nil {
		article = store.GetArticleById(articleId)
	}
	if article == nil {
		logger.Errorf("findArticleMustBeAdmin(): no article with article_id '%s'", idStr)
		httpErrorf(w, "invalid article")
	}
	return article
}

// app/delete?article_id=${id}
func handleAppDelete(w http.ResponseWriter, r *http.Request) {
	logger.Notice("handleAppDelete()")
	article := findArticleMustBeAdmin(w, r)
	if article == nil {
		return
	}

	if article.IsDeleted {
		logger.Errorf("handleAppDelete(): article %d already deleted", article.Id)
		httpErrorf(w, "article already deleted")
		return
	}

	article.IsDeleted = true
	store.UpdateArticle(article)
	clearArticlesCache()
	url := "/" + article.Permalink()
	http.Redirect(w, r, url, 301)
}

// app/undelete?article_id=${id}
func handleAppUndelete(w http.ResponseWriter, r *http.Request) {
	logger.Notice("handleAppUndelete()")
	article := findArticleMustBeAdmin(w, r)
	if article == nil {
		return
	}

	if !article.IsDeleted {
		logger.Errorf("handleAppUndelete(): article %d not deleted", article.Id)
		httpErrorf(w, "article not deleted")
		return
	}

	article.IsDeleted = false
	store.UpdateArticle(article)
	clearArticlesCache()
	url := "/" + article.Permalink()
	http.Redirect(w, r, url, 301)
}

// app/showdeleted
func handleAppShowDeleted(w http.ResponseWriter, r *http.Request) {
	logger.Notice("handleAppShowDeleted()")
	isAdmin := IsAdmin(r)
	if !isAdmin {
		http404(w, r)
		return
	}
	articles := make([]*Article, 0)
	for _, a := range getCachedArticles(isAdmin) {
		if a.IsDeleted {
			articles = append(articles, a)
		}
	}
	showArchiveArticles(w, r, articles, "")
}

// app/showprivate
func handleAppShowPrivate(w http.ResponseWriter, r *http.Request) {
	logger.Notice("handleAppShowPrivate()")
	isAdmin := IsAdmin(r)
	if !isAdmin {
		http404(w, r)
		return
	}
	articles := make([]*Article, 0)
	for _, a := range getCachedArticles(isAdmin) {
		if a.IsPrivate {
			articles = append(articles, a)
		}
	}
	showArchiveArticles(w, r, articles, "")
}
