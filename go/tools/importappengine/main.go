package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var dataDir = ""
var srcDataDir = filepath.Join("..", "..", "blogimported")
var dstDataDir = filepath.Join("..", "..", "blogdata")

const (
	FormatHtml     = 0
	FormatTextile  = 1
	FormatMarkdown = 2
	FormatText     = 3
)

type Text struct {
	Id        int
	CreatedOn time.Time
	Format    int
	Sha1Str   string
	Sha1      [20]byte
}

var newlines = []byte{'\n', '\n'}
var newline = []byte{'\n'}

func remSep(s string) string {
	return strings.Replace(s, "|", "", -1)
}

// "2006-06-05 17:06:34"
func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		log.Fatalf("failed to parse date %s, err: %s", s, err.Error())
	}
	return t
}

func parseText(d []byte) *Text {
	parts := bytes.Split(d, newline)
	text := &Text{}
	var err error
	for _, p := range parts {
		lp := bytes.SplitN(p, []byte{':', ' '}, 2)
		name := string(lp[0])
		val := string(lp[1])
		if name == "I" {
			if text.Id, err = strconv.Atoi(val); err != nil {
				log.Fatalf("invalid I val: '%s', err: %s\n", val, err.Error())
			}
		} else if name == "M" {
			text.Sha1Str = val
			sha1, err := hex.DecodeString(val)
			if err != nil || len(sha1) != 20 {
				log.Fatalf("error decoding M")
			}
			copy(text.Sha1[:], sha1)
		} else if name == "On" {
			text.CreatedOn = parseTime(val)
		} else if name == "F" {
			if val == "html" {
				text.Format = FormatHtml
			} else if val == "text" {
				text.Format = FormatText
			} else if val == "textile" {
				text.Format = FormatTextile
			} else if val == "markdown" {
				text.Format = FormatMarkdown
			} else {
				log.Fatalf("Unknown F val: '%s'\n", val)
			}
		} else {
			log.Fatalf("Unknown field: '%s'\n", name)
		}
	}
	return text
}

func parseTexts(d []byte) []*Text {
	texts := make([]*Text, 0)
	for len(d) > 0 {
		idx := bytes.Index(d, newlines)
		if idx == -1 {
			break
		}
		text := parseText(d[:idx])
		texts = append(texts, text)
		d = d[idx+2:]
	}
	return texts
}

func loadTexts() []*Text {
	filePath := filepath.Join(srcDataDir, "texts.txt")
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open %s, error: %s", filePath, err.Error())
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("loadTexts(): ioutil.ReadAll() failed with error: %s", err.Error())
	}
	return parseTexts(data)
}

func serText(t *Text) string {
	s1 := fmt.Sprintf("%d", t.CreatedOn.Unix())
	s2 := base64.StdEncoding.EncodeToString(t.Sha1[:])
	s2 = s2[:len(s2)-1] // remove '=' from the end
	return fmt.Sprintf("T%d|%s|%d|%s\n", t.Id, s1, t.Format, s2)
}

func serAll(texts []*Text) []string {
	res := make([]string, 0)
	for _, t := range texts {
		res = append(res, serText(t))
	}
	return res
}

func blobPath(dir, sha1 string) string {
	d1 := sha1[:2]
	d2 := sha1[2:4]
	return filepath.Join(dir, "blobs", d1, d2, sha1)
}

func copyBlobs(texts []*Text) error {
	for _, t := range texts {
		sha1 := t.Sha1Str
		srcPath := blobPath(srcDataDir, sha1)
		dstPath := blobPath(dstDataDir, sha1)
		if !PathExists(srcPath) {
			panic("srcPath doesn't exist")
		}
		if !PathExists(dstPath) {
			if err := CreateDirIfNotExists(filepath.Dir(dstPath)); err != nil {
				panic("failed to create dir for dstPath")
			}
			if err := CopyFile(dstPath, srcPath); err != nil {
				fmt.Printf("CopyFile('%s', '%s') failed with %s", dstPath, srcPath, err)
				return err
			}
			fmt.Sprintf("%s=>%s\n", srcPath, dstPath)
		}
	}
	return nil
}

func saveConverted(strs []string) {
	f, err := os.Create(filepath.Join(dstDataDir, "blogdata.txt"))
	if err != nil {
		log.Fatalf("os.Create() failed with %s", err.Error())
	}
	defer f.Close()
	for _, s := range strs {
		_, err = f.WriteString(s)
		if err != nil {
			log.Fatalf("WriteFile() failed with %s", err.Error())
		}
	}

}

func main() {
	if !PathExists(srcDataDir) {
		panic("srcDataDir doesn't exist")
	}
	if !PathExists(dstDataDir) {
		panic("dstDataDir doesn't exist")
	}
	texts := loadTexts()
	saveConverted(serAll(texts))
	if err := copyBlobs(texts); err != nil {
		panic("copyBlobs() failed")
	}
	fmt.Printf("%d texts\n", len(texts))
}
