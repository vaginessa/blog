package main

import (
	"fmt"
	"os"

	"github.com/kjk/u"
)

var (
	store2Rewrite *Store2
)

func RewriteStore(dataDir string) {
	var err error
	fmt.Printf("RewriteStore(%q)\n", dataDir)
	blobsBasePath := store2BlobsBasePath(dataDir)
	idxPath := blobsBasePath + "_idx.txt"

	if u.PathExists(idxPath) {
		fmt.Printf("RewriteStore: not rewriting because %q already exists\n", idxPath)
		return
	}

	// delete store data file
	os.Remove(store2Path(dataDir))

	// delete contentstore files
	os.Remove(idxPath)
	i := 0
	for {
		path := blobsBasePath + fmt.Sprintf("_%d.txt", i)
		if !u.PathExists(path) {
			break
		}
		os.Remove(path)
		i += 1
	}
	store2Rewrite, err = NewStore2(dataDir)
	panicif(err != nil, "NewStore2(%q) failed with %q", dataDir, err)

	store, err := NewStore(dataDir)
	panicif(err != nil, "NewStore(%q) failed with %q", dataDir, err)

	store2Rewrite.Close()
	store2Rewrite = nil
	store.Close()
}
