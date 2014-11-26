package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kjk/contentstore"
	"github.com/kjk/u"
)

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

func main() {
	d := u.ExpandTildeInPath("~/data/blog/data")
	path := filepath.Join(d, "blogdata2.txt")
	u.PanicIf(!u.PathExists(path))
	_, err := contentstore.NewWithLimit(filepath.Join(d, "blogblobs"), 4*1024*1024)
	u.PanicIfErr(err)
	readExistingBlogData(path)
	fmt.Printf("%d articles, %d versions\n", len(articles), len(texts))
}
