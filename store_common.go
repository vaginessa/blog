package main

import (
	"fmt"
	"strings"
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
