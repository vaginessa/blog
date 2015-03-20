package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kjk/u"
)

const (
	div = "--------------\n"
)

func isTextileFile(s string) bool {
	return strings.HasSuffix(s, ".textile")
}

func isHtmlFile(s string) bool {
	return strings.HasSuffix(s, ".html")
}

func shouldConvert(s string) bool {
	return isTextileFile(s) || isHtmlFile(s)
}

func getFilesToConvert(dir string) []string {
	res := make([]string, 0)
	dirsToVisit := []string{dir}
	for len(dirsToVisit) > 0 {
		dir := dirsToVisit[0]
		dirsToVisit = dirsToVisit[1:]
		entries, err := ioutil.ReadDir(dir)
		u.PanicIfErr(err)
		for _, fi := range entries {
			name := fi.Name()
			if fi.IsDir() {
				path := filepath.Join(dir, name)
				dirsToVisit = append(dirsToVisit, path)
			} else {
				if shouldConvert(name) {
					path := filepath.Join(dir, name)
					res = append(res, path)
				}
			}
		}
	}
	return res
}

func runCmd(cmdName string, args ...string) {
	cmd := exec.Command(cmdName, args...)
	err := cmd.Run()
	u.PanicIfErr(err)
}

func checkPandoc() {
	runCmd("pandoc", "--help")
}

func splitFile(path string) (string, string) {
	d, err := ioutil.ReadFile(path)
	u.PanicIfErr(err)
	s := string(d)
	s = strings.Replace(s, "\r\n", "\n", -1)
	s = strings.Replace(s, "\r", "\n", -1)
	idx := strings.Index(s, "----------")
	u.PanicIf(idx == -1, "idx == -1")
	hdr := s[:idx]
	hdr = strings.Replace(hdr, "Html", "Markdown", -1)
	hdr = strings.Replace(hdr, "Textile", "Markdown", -1)
	body := s[idx:]
	idx = strings.Index(body, "\n")
	u.PanicIf(idx == -1, "idx == -1")
	body = body[idx+1:]
	return hdr, body
}

func gitRename(path string) {
	ext := filepath.Ext(path)
	if ext == ".md" {
		return
	}
	dstPath := path[:len(path)-len(ext)]
	dstPath += ".md"
	//fmt.Printf("dst path: %s\n", dstPath)
	runCmd("git", "mv", path, dstPath)
}

func convertWithPandoc(path string) {
	var from string
	if isTextileFile(path) {
		from = "textile"
	} else if isHtmlFile(path) {
		from = "html"
	} else {
		panic("unknown format")
	}
	hdr, body := splitFile(path)
	pathTmp := path + ".tmp.markdown"
	err := ioutil.WriteFile(pathTmp, []byte(body), 0755)
	u.PanicIfErr(err)
	runCmd("pandoc", "-f", from, "-t", "markdown", "-o", pathTmp, pathTmp)
	converted, err := ioutil.ReadFile(pathTmp)
	u.PanicIfErr(err)
	f, err := os.Create(path)
	u.PanicIfErr(err)
	_, err = f.WriteString(hdr)
	u.PanicIfErr(err)
	_, err = f.WriteString(div)
	u.PanicIfErr(err)
	_, err = f.Write(converted)
	u.PanicIfErr(err)
	f.Close()
	err = os.Remove(pathTmp)
	u.PanicIfErr(err)
}

func main() {
	checkPandoc()
	files := getFilesToConvert("blog_posts")
	for _, path := range files {
		fmt.Printf("renaming: %s\n", path)
		//convertWithPandoc(path)
		gitRename(path)
	}
}
