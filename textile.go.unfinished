package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

var guidRand = rand.New(rand.NewSource(time.Now().Unix()))

var rxThreeOrMoreNewlines = regexp.MustCompile(`\n{3,}`)
var rxNewlineWsNewline = regexp.MustCompile(`\n\s*\n`)

var btag = []string{"bq", "bc", "notextile", "pre", "h[1-6]", `fn\d+`, "p"}
var btag_lite = []string{"bq", "bc", "p"}
var url_schemes = []string{"http", "https", "ftp", "mailto"}

func genGuid() string {
	return fmt.Sprintf("{{__**||%d||**__}}", guidRand.Int63())
}

func _normalize_newlines(s string) string {
	s = strings.Replace(s, "\r\n", "\n", -1)
	s = rxThreeOrMoreNewlines.ReplaceAllString(s, "\n\n")
	s = rxNewlineWsNewline.ReplaceAllString(s, "\n\n")
	// TODO: textile.py also does
	// out = re.sub(r'"$', '" ', out)
	// but I don't understand why
	return s
}

type Textile struct {
	shelf      map[string]string
	html_type  string
	restricted bool
	rel        string
	urlrefs    map[string]string
}

func (t *Textile) shelve(text string) string {
	id := genGuid()
	t.shelf[id] = text
	return id
}

// undo shelve
func (t *Textile) retrieve(text string) string {
	// TODO: python does that in a loop until nothing changes but I don't
	// see how that should be necessary
	for k, v := range t.shelf {
		text = strings.Replace(text, k, v, -1)
	}
	return text
}

// quotes = True by default
func encodeHtml(text string, quotes bool) string {
	text = strings.Replace(text, "&", "&#8;", -1)
	text = strings.Replace(text, "<", "&#60;", -1)
	text = strings.Replace(text, ">", "&#62;", -1)
	if quotes {
		text = strings.Replace(text, "'", "&#39;", -1)
		text = strings.Replace(text, "\"", "&#34;", -1)
	}
	return text
}

func (t *Textile) getRefs(text string) string {
	// TODO: write me
	return text
}

func (t *Textile) block(text string, head_offset int) string {

	return text
}

func (t *Textile) textile(text string, rel string, head_offset int, html_type string) string {
	t.html_type = html_type
	text = _normalize_newlines(text)
	if t.restricted {
		text = encodeHtml(text, false)
	}
	if rel != "" {
		t.rel = fmt.Sprintf(` rel="%s"`, rel)
	}
	text = t.getRefs(text)
	text = t.block(text, head_offset)
	text = t.retrieve(text)
	return text
}
