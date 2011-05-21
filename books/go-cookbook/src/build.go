package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	//"template"
)

func parse(r io.Reader) {
	buf := make([]byte, 128)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return
	}
}

func main() {
	dstdir := filepath.Join("..", "..", "www", "go-cookbook")
	_, err := os.Stat(dstdir)
	if err != nil {
		fmt.Printf("Dest dir doesn't exist, creating\n")
		os.MkdirAll(dstdir, 0666)
	}
	srcname := "book.txt"
	src, err := os.Open(srcname)
	if src == nil {
		fmt.Printf("Can't open '%s'\n", srcname)
		os.Exit(1)
	}
	defer src.Close()

}
