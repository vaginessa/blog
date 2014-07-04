package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	_ "errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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
var formatNames []string = []string{"Html", "Textile", "Markdown", "Text"}

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

type Store struct {
	sync.Mutex
	dataDir            string
	texts              []Text
	articles           []Article
	articleIdToArticle map[int]*Article
	dataFile           *os.File
	// cached data, returning full objects, not just pointers, to make them
	// read-only and therefore thread safe
	articlesCacheId int // increment when we do something that changes articles
	articlesCache   []Article
}

type ArticlesByTime []Article

func (s ArticlesByTime) Len() int {
	return len(s)
}

func (s ArticlesByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ArticlesByTime) Less(i, j int) bool {
	return s[i].PublishedOn.Before(s[j].PublishedOn)
}

func (a *Article) Permalink() string {
	return "article/" + ShortenId(a.Id) + "/" + Urlify(a.Title) + ".html"
}

func (a *Article) CurrVersion() *Text {
	vers := a.Versions
	return vers[len(vers)-1]
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

func (a *Article) FormatName() string {
	return formatNames[a.CurrVersion().Format]
}

func (a *Article) TagsDisplay() template.HTML {
	n := len(a.Tags)
	if n == 0 {
		return ""
	}
	arr := make([]string, n, n)
	for i, t := range a.Tags {
		arr[i] = urlForTag(t)
	}
	s := strings.Join(arr, ", ")
	return template.HTML(s)
}

func validFormat(format int) bool {
	return format >= FormatFirst && format <= FormatLast
}

// space saving: false values are empty strings, true is "1"
func strToBool(s string) bool {
	if s == "" {
		return false
	}
	if s == "1" {
		return true
	}
	panic("invalid bool string")
}

// space saving: false values are empty strings, true is "1"
func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return ""
}

func remSep(s string) string {
	return strings.Replace(s, "|", "", -1)
}

// parse:
// T1|1234860514|0|OiKDjvc+iyv4UXxVxLO91ozXwaU
func (s *Store) parseText(line []byte) {
	parts := strings.Split(string(line[1:]), "|")
	if len(parts) != 4 {
		panic("len(parts) != 4")
	}
	idStr := parts[0]
	createdOnSecondsStr := parts[1]
	formatStr := parts[2]
	msgSha1b64 := parts[3] + "="

	id, err := strconv.Atoi(idStr)
	if err != nil {
		panic("idStr not a number")
	}
	if id != len(s.texts) {
		panic("parseText(): invalid Text id")
	}

	createdOnSeconds, err := strconv.Atoi(createdOnSecondsStr)
	if err != nil {
		panic("createdOnSeconds not a number")
	}
	createdOn := time.Unix(int64(createdOnSeconds), 0)

	format, err := strconv.Atoi(formatStr)
	panicif(err != nil || !validFormat(format), "%s is not a valid format", formatStr)

	msgSha1, err := base64.StdEncoding.DecodeString(msgSha1b64)
	if err != nil {
		panic("msgSha1b64 not valid base64")
	}
	if len(msgSha1) != 20 {
		panic("len(msgSha1) != 20")
	}

	t := Text{
		Id:        id,
		CreatedOn: createdOn,
		Format:    format,
	}
	copy(t.Sha1[:], msgSha1)
	if !s.MessageFileExists(t.Sha1[:]) {
		panic("message file doesn't exist")
	}

	s.texts = append(s.texts, t)

	if store2Rewrite != nil {
		// TODO: read message data
		//fmt.Printf("Writing text id: %d sha1: %s\n", id, msgSha1b64)
		path := s.MessageFilePath(t.Sha1[:])
		d, err := ioutil.ReadFile(path)
		panicif(err != nil, "ReadFile(%q) failed with %q", path, err)
		_, err = store2Rewrite.CreateNewTextWithTime(format, string(d), createdOn)
		panicif(err != nil, "CreateNewTextWithTime() failed with %q", err)
	}
}

// parse:
// A582|$time|$title|$isPublic|$isDeleted|$tags|$versions
func (s *Store) parseArticle(line []byte) {
	parts := strings.Split(string(line[1:]), "|")
	if len(parts) != 7 {
		panic("len(parts) != 7")
	}
	idStr := parts[0]
	publishedOnStr := parts[1]
	title := parts[2]
	isPrivateStr := parts[3]
	isDeletedStr := parts[4]
	tagsStr := parts[5]
	versionIdsStr := parts[6]

	articleId, err := strconv.Atoi(idStr)
	if err != nil {
		panic("idStr not a number")
	}

	publishedOnSeconds, err := strconv.Atoi(publishedOnStr)
	if err != nil {
		panic("publishedOnSeconds not a number")
	}
	publishedOn := time.Unix(int64(publishedOnSeconds), 0)

	isPrivate := strToBool(isPrivateStr)
	isDeleted := strToBool(isDeletedStr)
	var tags []string
	if tagsStr == "" {
		tags = make([]string, 0)
	} else {
		tags = strings.Split(tagsStr, ",")
	}

	versionsStr := strings.Split(versionIdsStr, ",")
	nVersions := len(versionsStr)
	if nVersions == 0 {
		panic("We need some versions")
	}

	versions := make([]*Text, nVersions, nVersions)
	for i, verStr := range versionsStr {
		textId, err := strconv.Atoi(verStr)
		if err != nil {
			panic("verStr not a number")
		}
		if textId > len(s.texts) {
			panic("non-existent verStr")
		} else {
			versions[i] = &s.texts[textId]
		}
	}

	var a *Article
	var existingArticle bool
	if a, existingArticle = s.articleIdToArticle[articleId]; !existingArticle {
		a = &Article{Id: articleId}
	}

	a.PublishedOn = publishedOn
	a.IsPrivate = isPrivate
	a.IsDeleted = isDeleted
	a.Title = title
	a.Tags = tags
	a.Versions = versions

	if existingArticle {
		return
	}

	s.articles = append(s.articles, *a)
	s.articleIdToArticle[articleId] = &s.articles[len(s.articles)-1]
}

func (s *Store) readExistingBlogData(fileDataPath string) error {
	d, err := ioutil.ReadFile(fileDataPath)
	if err != nil {
		return err
	}

	for len(d) > 0 {
		idx := bytes.IndexByte(d, '\n')
		if -1 == idx {
			// TODO: this could happen if the last record was only
			// partially written. Should I just ignore it?
			panic("idx shouldn't be -1")
		}
		line := d[:idx]
		d = d[idx+1:]
		c := line[0]
		if c == 'T' {
			s.parseText(line)
		} else if c == 'A' {
			s.parseArticle(line)
		} else {
			panic("Unexpected line type")
		}
	}
	if store2Rewrite != nil {
		for _, a := range s.articles {
			store2Rewrite.writeArticleOld(&a)
			panicif(err != nil, "store2Rewrite.writeArticleOld() failed with %q", err)
		}
	}
	return nil
}

func NewStore(dataDir string) (*Store, error) {
	dataFilePath := filepath.Join(dataDir, "data", "blogdata.txt")
	store := &Store{
		dataDir:            dataDir,
		texts:              make([]Text, 0),
		articles:           make([]Article, 0),
		articleIdToArticle: make(map[int]*Article),
		articlesCacheId:    1,
	}
	var err error
	if u.PathExists(dataFilePath) {
		err = store.readExistingBlogData(dataFilePath)
		if err != nil {
			logger.Errorf("NewStore(): readExistingBlogData() failed with %s\n", err)
			return nil, err
		}
	} else {
		f, err := os.Create(dataFilePath)
		if err != nil {
			logger.Errorf("NewStore(): os.Create(%s) failed with %s", dataFilePath, err)
			return nil, err
		}
		f.Close()
	}
	store.dataFile, err = os.OpenFile(dataFilePath, os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		logger.Errorf("NewStore(): os.OpenFile(%s) failed with %s", dataFilePath, err)
		return nil, err
	}
	logger.Noticef("texts: %d, articles: %d", len(store.texts), len(store.articles))
	return store, nil
}

func (s *Store) ArticlesCount() int {
	s.Lock()
	defer s.Unlock()
	return len(s.articles)
}

func blobPath(dir, sha1 string) string {
	d1 := sha1[:2]
	d2 := sha1[2:4]
	return filepath.Join(dir, "blobs", d1, d2, sha1)
}

func (s *Store) MessageFilePath(sha1 []byte) string {
	sha1Str := hex.EncodeToString(sha1)
	return blobPath(s.dataDir, sha1Str)
}

func (s *Store) MessageFileExists(sha1 []byte) bool {
	p := s.MessageFilePath(sha1)
	return u.PathExists(p)
}

func (s *Store) appendString(str string) error {
	_, err := s.dataFile.WriteString(str)
	if err != nil {
		logger.Errorf("Store.appendString() error: %s\n", err)
	}
	return err
}

func (s *Store) writeMessageAsSha1(msg []byte, sha1 []byte) error {
	path := s.MessageFilePath(sha1)
	err := u.WriteBytesToFile(msg, path)
	if err != nil {
		logger.Errorf("Store.writeMessageAsSha1(): failed to write %s with error %s", path, err)
	}
	return err
}

func (s *Store) GetArticles(lastId int) (int, []Article) {
	s.Lock()
	defer s.Unlock()
	if s.articlesCache != nil && s.articlesCacheId == lastId {
		return s.articlesCacheId, s.articlesCache
	}

	n := len(s.articles)
	articles := make([]Article, n, n)
	for i, a := range s.articles {
		articles[i] = a
	}
	sort.Sort(ArticlesByTime(articles))
	s.articlesCache = articles
	return s.articlesCacheId, s.articlesCache
}

// TODO: those are not sorted by the real creation date. Good thing
// it's not used anymore (use articles_cache.go instead)
/*
func (s *Store) GetRecentArticles(max int, isAdmin bool) []*Article {
	s.Lock()
	defer s.Unlock()

	left := max
	res := make([]*Article, 0)
	idx := len(s.articles) - 1
	for left > 0 && idx >= 0 {
		a := &s.articles[idx]
		if (!a.IsPrivate && !a.IsDeleted) || isAdmin {
			res = append(res, a)
			left -= 1
		}
		idx -= 1
	}
	return res
}
*/

func (s *Store) GetArticleById(id int) *Article {
	s.Lock()
	defer s.Unlock()
	if article, ok := s.articleIdToArticle[id]; ok {
		return article
	}
	return nil
}

func (s *Store) newArticleId() int {
	id := 1
	for {
		if _, ok := s.articleIdToArticle[id]; !ok {
			return id
		}
		id += 1
	}
}

func (s *Store) addText(t Text) *Text {
	n := len(s.texts)
	t.Id = n
	s.texts = append(s.texts, t)
	return &s.texts[n]
}

func serText(t *Text) string {
	s1 := fmt.Sprintf("%d", t.CreatedOn.Unix())
	s2 := base64.StdEncoding.EncodeToString(t.Sha1[:])
	s2 = s2[:len(s2)-1] // remove '=' from the end
	return fmt.Sprintf("T%d|%s|%d|%s\n", t.Id, s1, t.Format, s2)
}

func (s *Store) CreateNewText(format int, txt string) (*Text, error) {
	panicif(!validFormat(format), "%d is not a valid fomrat", format)

	s.Lock()
	defer s.Unlock()

	data := []byte(txt)
	sha1 := u.Sha1OfBytes(data)
	if err := s.writeMessageAsSha1(data, sha1); err != nil {
		return nil, err
	}
	t := Text{
		CreatedOn: time.Now(),
		Format:    format,
	}
	copy(t.Sha1[:], sha1)
	if err := s.appendString(serText(&t)); err != nil {
		return nil, err
	}
	return s.addText(t), nil
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

func serArticle(a *Article) string {
	s1 := fmt.Sprintf("%d", a.Id)
	s2 := fmt.Sprintf("%d", a.PublishedOn.Unix())
	s3 := remSep(a.Title)
	s4 := boolToStr(a.IsPrivate)
	s5 := boolToStr(a.IsDeleted)
	s6 := serTags(a.Tags)
	nVers := len(a.Versions)
	vers := make([]string, nVers, nVers)
	for i, ver := range a.Versions {
		vers[i] = strconv.Itoa(ver.Id)
	}
	s7 := strings.Join(vers, ",")
	return fmt.Sprintf("A%s|%s|%s|%s|%s|%s|%s\n", s1, s2, s3, s4, s5, s6, s7)
}

func (s *Store) CreateOrUpdateArticle(article *Article) (*Article, error) {
	s.Lock()
	defer s.Unlock()

	newArticle := false
	if article.Id == 0 {
		article.Id = s.newArticleId()
		newArticle = true
	}
	articleStr := serArticle(article)
	if err := s.appendString(articleStr); err != nil {
		return nil, err
	}

	if newArticle {
		s.articles = append(s.articles, *article)
		article = &s.articles[len(s.articles)-1]
		s.articleIdToArticle[article.Id] = article
	}
	return article, nil
}

func (s *Store) UpdateArticle(article *Article) (*Article, error) {
	s.Lock()
	defer s.Unlock()

	tmp := s.articleIdToArticle[article.Id]
	if tmp != article {
		panic("invalid article object")
	}
	err := s.appendString(serArticle(article))
	return article, err
}

func (s *Store) Close() {
	if s.dataFile != nil {
		s.dataFile.Close()
	}
}
