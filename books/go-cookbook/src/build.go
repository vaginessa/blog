package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	dstpath := filepath.Join("..", "..", "www", "go-cookbook")
	_, err := os.Stat(dstpath)
	if err != nil {
		fmt.Printf("Directory doesn't exist, creating\n")
		os.MkdirAll(dstpath, 0666)
	}
}
