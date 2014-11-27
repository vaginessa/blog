package main

import (
	"net/http"
	"strconv"
	"strings"
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

func GetArticleVersionBody(bodyId string) (string, error) {
	msg, err := store.GetTextBody(bodyId)
	if err != nil {
		return "", err
	}
	return string(msg), nil
}

func findArticleMustBeAdmin(w http.ResponseWriter, r *http.Request) *Article2 {
	if !IsAdmin(r) {
		http.NotFound(w, r)
		return nil
	}

	var article *Article2
	idStr := getTrimmedFormValue(r, "article_id")
	if articleId, err := strconv.Atoi(idStr); err == nil {
		article = store.GetArticleById(articleId)
	}
	if article == nil {
		logger.Errorf("findArticleMustBeAdmin(): no article with article_id %q", idStr)
		httpErrorf(w, "invalid article")
	}
	return article
}
