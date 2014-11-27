package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	FormatHtml     = 0
	FormatTextile  = 1
	FormatMarkdown = 2
	FormatText     = 3

	FormatFirst   = 0
	FormatLast    = 3
	FormatUnknown = -1
)

type Text struct {
	Id        int
	CreatedOn time.Time
	Format    int
	Sha1      [20]byte
}

type Article struct {
	Id          int
	PublishedOn time.Time
	Title       string
	IsPrivate   bool
	IsDeleted   bool
	Tags        []string
	Versions    []*Text
}

// same format as Format* constants
var formatNames = []string{"Html", "Textile", "Markdown", "Text"}

func validFormat(format int) bool {
	return format >= FormatFirst && format <= FormatLast
}

func remSep(s string) string {
	return strings.Replace(s, "|", "", -1)
}

func urlForTag(tag string) string {
	// TODO: url-quote the first tag
	return fmt.Sprintf(`<a href="/tag/%s" class="taglink">%s</a>`, tag, tag)
}

func FormatNameToId(name string) int {
	for i, formatName := range formatNames {
		if strings.EqualFold(name, formatName) {
			return i
		}
	}
	return FormatUnknown
}

func joinStringsSanitized(arr []string, sep string) string {
	for i, s := range arr {
		// TODO: could also escape
		arr[i] = strings.Replace(s, sep, "", -1)
	}
	return strings.Join(arr, sep)
}

func serTags(tags []string) string {
	return joinStringsSanitized(tags, ",")
}

func deserTags(s string) []string {
	tags := strings.Split(s, ",")
	if len(tags) > 0 && tags[0] == "" {
		tags = tags[1:]
	}
	return tags
}

func deserVersions(s string) []string {
	return strings.Split(s, ",")
}
