package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kjk/contentstore"
	"github.com/kjk/u"
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
var formatExts = []string{".html", ".textile", ".md", ".txt"}

type Text struct {
	CreatedOn time.Time
	Text      string
	Format    int
}

type Article struct {
	PublishedOn time.Time
	Title       string
	IsPrivate   bool
	IsDeleted   bool
	Tags        []string
	Versions    []int
}

var (
	texts    map[int]*Text
	articles map[int]*Article
)

func init() {
	texts = make(map[int]*Text)
	articles = make(map[int]*Article)
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

// Store.Get(bodyId), bodyId == string

/* csv records:
t, $id, $createdOn, $format, $bodyId
a, $id, $publishedOn, $title, $flags, $tags, $versions
*/
func decodeRec(rec []string) {
	u.PanicIf(len(rec) < 5)

	id, err := strconv.Atoi(rec[1])
	u.PanicIfErr(err)
	timeSecs, err := strconv.ParseInt(rec[2], 10, 64)
	u.PanicIfErr(err)
	time := time.Unix(timeSecs, 0)

	if rec[0] == "t" {
		format, err := strconv.Atoi(rec[3])
		u.PanicIfErr(err)

		texts[id] = &Text{
			CreatedOn: time,
			Format:    format,
			Text:      rec[4],
		}
		return
	}

	if rec[0] == "a" {
		title := rec[3]
		var isPriv, isDel bool
		flags := rec[4]
		for _, flag := range flags {
			switch flag {
			case 'p':
				isPriv = true
			case 'd':
				isDel = true
			}
		}
		tags := deserTags(rec[5])
		versStr := deserVersions(rec[6])
		nVers := len(versStr)
		versions := make([]int, nVers, nVers)
		for i, ver := range versStr {
			textId, err := strconv.Atoi(ver)
			u.PanicIfErr(err)
			//u.PanicIf(textId > len(s.texts), "textId > len(s.texts) %d > %d", textId, len(s.texts))
			versions[i] = textId
		}
		a := &Article{
			Title:       title,
			PublishedOn: time,
			IsDeleted:   isDel,
			IsPrivate:   isPriv,
			Tags:        tags,
			Versions:    versions,
		}
		articles[id] = a
	}
}

func readExistingBlogData(fileDataPath string) {
	file, err := os.Open(fileDataPath)
	u.PanicIfErr(err)
	defer file.Close()
	csvReader := csv.NewReader(file)
	csvReader.Comma = ','
	csvReader.FieldsPerRecord = -1
	var rec []string
	for {
		if rec, err = csvReader.Read(); err != nil {
			break
		}
		decodeRec(rec)
	}
	u.PanicIf(err != io.EOF)
}

func sanitizeForFile(s string) string {
	var res []byte
	toRemove := "/\\#()[]{},?+.'\""
	var prev rune
	buf := make([]byte, 3)
	for _, c := range s {
		if strings.ContainsRune(toRemove, c) {
			continue
		}
		switch c {
		case ' ', '_':
			c = '-'
		}
		if c == prev {
			continue
		}
		prev = c
		n := utf8.EncodeRune(buf, c)
		for i := 0; i < n; i++ {
			res = append(res, buf[i])
		}
	}
	if len(res) > 32 {
		res = res[:32]
	}
	s = string(res)
	s = strings.Trim(s, "_- ")
	s = strings.ToLower(s)
	return s
}

func main() {
	d := u.ExpandTildeInPath("~/data/blog/data")
	path := filepath.Join(d, "blogdata2.txt")
	u.PanicIf(!u.PathExists(path))
	store, err := contentstore.NewWithLimit(filepath.Join(d, "blogblobs"), 4*1024*1024)
	u.PanicIfErr(err)
	readExistingBlogData(path)
	fmt.Printf("%d articles, %d versions\n", len(articles), len(texts))
	for id, article := range articles {
		title := article.Title
		isPriv := article.IsPrivate
		if isPriv {
			fmt.Printf("skipping private %s\n", title)
			continue
		}

		n := len(article.Versions)
		text := texts[article.Versions[n-1]]
		date := article.PublishedOn
		tags := article.Tags
		isDel := article.IsDeleted
		hdr := fmt.Sprintf("Id: %d\n", id)
		hdr += fmt.Sprintf("Title: %s\n", title)
		if isDel {
			hdr += fmt.Sprintf("Deleted: true\n")
		}
		if len(tags) > 0 {
			hdr += fmt.Sprintf("Tags: %s\n", strings.Join(tags, ","))
		}
		hdr += fmt.Sprintf("Date: %s\n", date.Format(time.RFC3339))
		hdr += fmt.Sprintf("Format: %s\n", formatNames[text.Format])

		dir := date.Format("2006-01")
		fileName := sanitizeForFile(title) + formatExts[text.Format]
		path := filepath.Join("blog_posts", dir, fileName)
		//fmt.Printf("Path: %s\n", path)
		hdr += "--------------\n"
		sha1 := text.Text
		body, err := store.Get(sha1)
		u.PanicIfErr(err)
		fmt.Print(hdr)
		all := hdr + string(body) + "\n"
		u.CreateDirForFileMust(path)
		ioutil.WriteFile(path, []byte(all), 0644)
	}
}
