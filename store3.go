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

type Text2 struct {
	Id        int
	CreatedOn time.Time
	Format    int
	BodyId    string
}

type Article2 struct {
	Id          int
	PublishedOn time.Time
	Title       string
	IsPrivate   bool
	IsDeleted   bool
	Tags        []string
	Versions    []*Text2
}

type Store3 struct {
	articles []*Article2
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

func readArticleMetadata(path string) (*Article2, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	a := &Article2{}
	a.Versions = make([]*Text2, 1, 1)
	a.Versions[0] = &Text2{}
	a.Versions[0].BodyId = path
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
			a.Versions[0].Id = id
		case "title":
			a.Title = v
		case "tags":
			a.Tags = parseTags(v)
		case "format":
			f := parseFormat(v)
			if f == FormatUnknown {
				return nil, fmt.Errorf("%q is not a valid format", v)
			}
			a.Versions[0].Format = f
		case "date":
			a.PublishedOn, err = parseDate(v)
			if err != nil {
				return nil, fmt.Errorf("%q is not a valid date", v)
			}
			a.Versions[0].CreatedOn = a.PublishedOn
		default:
			return nil, fmt.Errorf("Unexpected key: %q\n", k)
		}
	}
	return a, nil
}

func readArticlesMetadata() ([]*Article2, error) {
	timeStart := time.Now()
	walker := fs.Walk("blog_posts")
	res := make([]*Article2, 0)
	for walker.Step() {
		if walker.Err() != nil {
			fmt.Printf("walker.Err() failed with %s\n", walker.Err())
			return nil, walker.Err()
		}
		st := walker.Stat()
		if st.IsDir() {
			continue
		}
		path := walker.Path()
		//fmt.Printf("p: %s\n", path)
		a, err := readArticleMetadata(path)
		if err != nil {
			fmt.Printf("readArticleMetadata() failed with %s\n", err)
			return nil, err
		}
		//a.Path = path
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

func (s *Store3) GetArticles(lastId int) (int, []*Article2) {
	fmt.Printf("GetArticles: lastId: %d, nArticles: %d\n", lastId, len(s.articles))
	return 0, s.articles
}

func (s *Store3) GetTextBody(bodyId string) ([]byte, error) {
	fmt.Printf("GetTextBody: bodyId=%s\n", bodyId)
	//return s.Store.Get(bodyId)
	return nil, nil
}

func (s *Store3) GetArticleById(id int) *Article2 {
	fmt.Printf("GetArticleById: %d\n", id)
	for _, a := range s.articles {
		if a.Id == id {
			return a
		}
	}
	return nil
}

func (s *Store3) ArticlesCount() int {
	return len(s.articles)
}

func (a *Article2) CurrVersion() *Text2 {
	return a.Versions[0]
}

func (a *Article2) Permalink() string {
	return "article/" + ShortenId(a.Id) + "/" + Urlify(a.Title) + ".html"
}
