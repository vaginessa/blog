package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/kjk/contentstore"
)

/* csv records:
t, $id, $createdOn, $format, $sha1
a, $id, $publishedOn, $title, $isPrivate, $isDeleted, $tags, $versions
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
	filePath string
	file     *os.File
	w        *csv.Writer
	texts    []*Text2
	articles []*Article2
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

func (s *Store2) writeCsv(rec []string) error {
	recs := [][]string{rec}
	return s.w.WriteAll(recs)
}

func (s *Store2) writeTextRec(t *Text2) error {
	timeStr := strconv.FormatInt(t.CreatedOn.Unix(), 10)
	formatStr := strconv.Itoa(t.Format)
	rec := []string{recIdText, strconv.Itoa(t.Id), timeStr, formatStr, t.BodyId}
	return s.writeCsv(rec)
}

func NewStore2(dataDir string) (*Store2, error) {
	dataFilePath := store2Path(dataDir)
	blobsBasePath := store2BlobsBasePath(dataDir)
	contentStore, err := contentstore.NewWithLimit(blobsBasePath, 4*1024*1024)
	if err != nil {
		return nil, err
	}

	s := &Store2{}
	s.Store = contentStore
	if s.file, s.w, err = openCsv(dataFilePath); err != nil {
		return nil, err
	}
	s.filePath = dataFilePath
	return s, nil
}

func (s *Store2) nextTextId() int {
	return len(s.texts)
}

func (s *Store2) addText(t *Text2) {
	n := len(s.texts)
	t.Id = n
	s.texts = append(s.texts, t)
}

func (s *Store2) CreateNewText(format int, txt string) (*Text2, error) {
	return s.CreateNewTextWithTime(format, txt, time.Now())
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
