package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kjk/contentstore"
	"github.com/kjk/u"
)

/* csv records:
t, $id, $createdOn, $format, $sha1
a, $id, $publishedOn, $title, $flags, $tags, $versions
*/

const (
	recIdText    = "t"
	recIdArticle = "a"
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

type Store2 struct {
	sync.Mutex
	*contentstore.Store
	filePath           string
	file               *os.File
	w                  *csv.Writer
	texts              []*Text2
	articles           []*Article2
	articleIdToArticle map[int]*Article2
	// cached data, returning full objects, not just pointers, to make them
	// read-only and therefore thread safe
	articlesCacheId int // increment when we do something that changes articles
	articlesCache   []*Article2
}

func (s *Store2) GetTextBody(bodyId string) ([]byte, error) {
	return s.Store.Get(bodyId)
}

func store2Path(dir string) string {
	return filepath.Join(dir, "data", "blogdata2.txt")
}

func store2BlobsBasePath(dir string) string {
	return filepath.Join(dir, "data", "blogblobs")
}

func openCsv(path string) (*os.File, *csv.Writer, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, nil, err
	} else {
		return file, csv.NewWriter(file), nil
	}
}

type Articles2ByTime []*Article2

func (s Articles2ByTime) Len() int {
	return len(s)
}

func (s Articles2ByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Articles2ByTime) Less(i, j int) bool {
	return s[i].PublishedOn.Before(s[j].PublishedOn)
}

func (a *Article2) Permalink() string {
	return "article/" + ShortenId(a.Id) + "/" + Urlify(a.Title) + ".html"
}

func (a *Article2) CurrVersion() *Text2 {
	vers := a.Versions
	return vers[len(vers)-1]
}

func (a *Article2) FormatName() string {
	return formatNames[a.CurrVersion().Format]
}

func (a *Article2) TagsDisplay() template.HTML {
	arr := make([]string, 0)
	for _, tag := range a.Tags {
		arr = append(arr, urlForTag(tag))
	}
	s := strings.Join(arr, ", ")
	return template.HTML(s)
}

func (s *Store2) writeCsv(rec []string) error {
	recs := [][]string{rec}
	return s.w.WriteAll(recs)
}

// t, $id, $createdOn, $format, $sha1
func (s *Store2) writeTextRec(t *Text2) error {
	timeStr := strconv.FormatInt(t.CreatedOn.Unix(), 10)
	formatStr := strconv.Itoa(t.Format)
	rec := []string{recIdText, strconv.Itoa(t.Id), timeStr, formatStr, t.BodyId}
	return s.writeCsv(rec)
}

// a, $id, $publishedOn, $title, $flags, $tags, $versions
func (s *Store2) writeArticleRec(a *Article2) error {
	timeStr := strconv.FormatInt(a.PublishedOn.Unix(), 10)
	idStr := strconv.Itoa(a.Id)
	flags := ""
	if a.IsPrivate {
		flags += "p"
	}
	if a.IsDeleted {
		flags += "d"
	}
	tags := serTags(a.Tags)
	nVers := len(a.Versions)
	vers := make([]string, nVers, nVers)
	for i, ver := range a.Versions {
		vers[i] = strconv.Itoa(ver.Id)
	}
	versions := strings.Join(vers, ",")
	rec := []string{recIdArticle, idStr, timeStr, a.Title, flags, tags, versions}
	return s.writeCsv(rec)
}

// only needed for rewrite
func (s *Store2) writeArticleOld(a *Article) error {
	aNew := &Article2{
		Id:          a.Id,
		PublishedOn: a.PublishedOn,
		Title:       a.Title,
		IsPrivate:   a.IsPrivate,
		IsDeleted:   a.IsDeleted,
		Tags:        a.Tags,
		Versions:    make([]*Text2, 0),
	}
	for _, ver := range a.Versions {
		verId := ver.Id
		verNew := s.texts[verId]
		aNew.Versions = append(aNew.Versions, verNew)
	}
	return s.writeArticleRec(aNew)
}

func NewStore2(dataDir string) (*Store2, error) {
	dataFilePath := store2Path(dataDir)
	s := &Store2{
		texts:              make([]*Text2, 0),
		articles:           make([]*Article2, 0),
		articleIdToArticle: make(map[int]*Article2),
		articlesCacheId:    1,
	}
	blobsBasePath := store2BlobsBasePath(dataDir)
	contentStore, err := contentstore.NewWithLimit(blobsBasePath, 4*1024*1024)
	if err != nil {
		return nil, err
	}
	s.Store = contentStore

	if u.PathExists(dataFilePath) {
		err = s.readExistingBlogData(dataFilePath)
		if err != nil {
			logger.Errorf("NewStore(): readExistingBlogData() failed with %s\n", err)
			return nil, err
		}
	}

	if s.file, s.w, err = openCsv(dataFilePath); err != nil {
		return nil, err
	}
	s.filePath = dataFilePath
	return s, nil
}

/* csv records:
t, $id, $createdOn, $format, $bodyId
a, $id, $publishedOn, $title, $flags, $tags, $versions
*/
func (s *Store2) decodeRec(rec []string) error {
	if len(rec) < 5 {
		return fmt.Errorf("rec of invalid len %d", len(rec))
	}

	id, err := strconv.Atoi(rec[1])
	if err != nil {
		return err
	}
	timeSecs, err := strconv.ParseInt(rec[2], 10, 64)
	if err != nil {
		return err
	}
	time := time.Unix(timeSecs, 0)

	if rec[0] == recIdText {
		format, err := strconv.Atoi(rec[3])
		if err != nil {
			return err
		}
		t := &Text2{
			Id:        id,
			CreatedOn: time,
			Format:    format,
			BodyId:    rec[4],
		}
		panicif(t.Id != len(s.texts), "t.Id != len(s.texts) (%d != %d)", t.Id, len(s.texts))
		s.texts = append(s.texts, t)
	} else if rec[0] == recIdArticle {
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
		versions := make([]*Text2, nVers, nVers)
		for i, ver := range versStr {
			textId, err := strconv.Atoi(ver)
			if err != nil {
				return err
			}
			panicif(textId > len(s.texts), "textId > len(s.texts) %d > %d", textid, len(s.texts))
			versions[i] = s.texts[textId]
		}
		a := &Article2{
			Id:          id,
			Title:       title,
			PublishedOn: time,
			IsDeleted:   isDel,
			IsPrivate:   isPriv,
			Tags:        tags,
			Versions:    versions,
		}
		s.articles = append(s.articles, a)
		s.articleIdToArticle[a.Id] = a
	} else {
		return fmt.Errorf("invalid rec[0] = %s", rec[0])
	}
	return nil
}

func (s *Store2) readExistingBlogData(fileDataPath string) error {
	file, err := os.Open(fileDataPath)
	if err != nil {
		return err
	}
	defer file.Close()
	csvReader := csv.NewReader(file)
	csvReader.Comma = ','
	csvReader.FieldsPerRecord = -1
	var rec []string
	for {
		if rec, err = csvReader.Read(); err != nil {
			break
		}
		if err = s.decodeRec(rec); err != nil {
			break
		}
	}
	if err == io.EOF {
		err = nil
	}
	return nil
}

func (s *Store2) nextTextId() int {
	return len(s.texts)
}

func (s *Store2) addText(t *Text2) {
	n := len(s.texts)
	t.Id = n
	s.texts = append(s.texts, t)
}

func (s *Store2) ArticlesCount() int {
	s.Lock()
	defer s.Unlock()
	return len(s.articles)
}

func (s *Store2) CreateNewText(format int, txt string) (*Text2, error) {
	return s.CreateNewTextWithTime(format, txt, time.Now())
}

func (s *Store2) GetArticles(lastId int) (int, []*Article2) {
	s.Lock()
	defer s.Unlock()
	if s.articlesCache != nil && s.articlesCacheId == lastId {
		return s.articlesCacheId, s.articlesCache
	}

	n := len(s.articles)
	articles := make([]*Article2, n, n)
	for i, a := range s.articles {
		articles[i] = a
	}
	sort.Sort(Articles2ByTime(articles))
	s.articlesCache = articles
	return s.articlesCacheId, s.articlesCache
}

func (s *Store2) GetArticleById(id int) *Article2 {
	s.Lock()
	defer s.Unlock()
	if article, ok := s.articleIdToArticle[id]; ok {
		return article
	}
	return nil
}

func (s *Store2) newArticleId() int {
	id := 1
	for {
		if _, ok := s.articleIdToArticle[id]; !ok {
			return id
		}
		id += 1
	}
}

// TODO: only needed for rewrite
func (s *Store2) CreateNewTextWithTime(format int, txt string, createdOn time.Time) (*Text2, error) {
	panicif(!validFormat(format), "%d is not a valid format", format)
	s.Lock()
	defer s.Unlock()

	data := []byte(txt)
	bodyId, err := s.Store.Put(data)
	if err != nil {
		return nil, err
	}
	t := &Text2{
		Id:        s.nextTextId(),
		CreatedOn: createdOn,
		Format:    format,
		BodyId:    bodyId,
	}
	if err = s.writeTextRec(t); err != nil {
		return nil, err
	}
	s.addText(t)
	return t, nil
}

func (s *Store2) CreateOrUpdateArticle(article *Article2) (*Article2, error) {
	s.Lock()
	defer s.Unlock()

	newArticle := false
	if article.Id == 0 {
		article.Id = s.newArticleId()
		newArticle = true
	}
	if err := s.writeArticleRec(article); err != nil {
		return nil, err
	}

	if newArticle {
		s.articles = append(s.articles, article)
		s.articleIdToArticle[article.Id] = article
	}
	return article, nil
}

func (s *Store2) UpdateArticle(article *Article2) (*Article2, error) {
	s.Lock()
	defer s.Unlock()

	tmp := s.articleIdToArticle[article.Id]
	if tmp != article {
		panic("invalid article object")
	}
	err := s.writeArticleRec(article)
	return article, err
}

func (s *Store2) Close() {
	if s.Store != nil {
		s.Store.Close()
		s.Store = nil
	}
	if s.file != nil {
		s.file.Close()
		s.file = nil
	}
}
