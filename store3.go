package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kr/fs"
)

type Article3 struct {
	Id        int
	Title     string
	Format    int
	CreatedOn time.Time
	Tags      []string
	Path      string
}

type Store3 struct {
	articles []*Article3
}

func isSepLine(s string) bool {
	return strings.HasPrefix(s, "-----")
}

func parseTags(s string) []string {
	tags := strings.Split(s, ",")
	for i, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)
		tags[i] = tag
	}
	return tags
}

func parseFormat(s string) int {
	s = strings.ToLower(s)
	switch s {
	case "html":
		return FormatHtml
	case "textile":
		return FormatTextile
	case "markdown", "md":
		return FormatMarkdown
	case "text":
		return FormatText
	default:
		return FormatUnknown
	}
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(s, "2006-01-02")
	if err == nil {
		return t, nil
	}
	// TODO: more formats?
	return time.Now(), err
}

func readArticleMetadata(path string) (*Article3, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	a := &Article3{}
	r := bufio.NewReader(f)
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		l = strings.TrimSpace(l)
		if isSepLine(l) {
			break
		}
		parts := strings.SplitN(l, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("Unexpected line: %q\n", l)
		}
		k := strings.ToLower(parts[0])
		v := strings.TrimSpace(parts[1])
		switch k {
		case "deleted":
			continue // skip deleted articles
		case "id":
			id, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("%q is not a valid id (not a number)", v)
			}
			a.Id = id
		case "title":
			a.Title = v
		case "tags":
			a.Tags = parseTags(v)
		case "format":
			a.Format = parseFormat(v)
			if a.Format == FormatUnknown {
				return nil, fmt.Errorf("%q is not a valid format", v)
			}
		case "date":
			a.CreatedOn, err = parseDate(v)
			if err != nil {
				return nil, fmt.Errorf("%q is not a valid date", v)
			}
		default:
			return nil, fmt.Errorf("Unexpected key: %q\n", k)
		}
	}
	return a, nil
}

func readArticlesMetadata() ([]*Article3, error) {
	timeStart := time.Now()
	walker := fs.Walk("blog_post")
	res := make([]*Article3, 0)
	for walker.Step() {
		stat := walker.Stat()
		if stat.IsDir() {
			continue
		}
		path := walker.Path()
		//fmt.Printf("p: %s\n", path)
		a, err := readArticleMetadata(path)
		if err != nil {
			fmt.Printf("readArticleMetadata() failed with %s\n", err)
			return nil, err
		}
		a.Path = path
		res = append(res, a)
	}
	fmt.Printf("read %d articles in %s\n", len(res), time.Since(timeStart))
	return res, nil
}

func NewStore3() (*Store3, error) {
	articles, err := readArticlesMetadata()
	if err != nil {
		return nil, err
	}
	return &Store3{articles: articles}, nil
}

func (s *Store3) CreateNewText(format int, txt string) (*Text2, error) {
	panic("NYI")
}

func (s *Store3) GetArticles(lastId int) (int, []*Article2) {
	return 0, nil
}
