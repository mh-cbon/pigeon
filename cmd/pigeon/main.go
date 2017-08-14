package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kaneshin/pigeon"
	"github.com/kaneshin/pigeon/cache"
	homedir "github.com/mitchellh/go-homedir"
)

func main() {
	// Parse arguments to run this function.
	detects := DetectionsParse(os.Args[1:])

	if args := detects.Args(); len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		detects.Usage()
		os.Exit(1)
	}

	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Unable to get yur home directory: %v\n", err)
	}

	home = filepath.Join(home, ".gvision-cache")
	// store := cache.FsStore{Path:home , Expire: time.Minute * 30}
	store := &cache.SQLStore{Db: cache.Sqlite(home), Expire: time.Minute * 30}
	if err2 := store.Init(); err2 != nil {
		log.Fatalf("Failed to intialize the cache store: %v\n", err2)
	}

	// Initialize vision service by a credentials json.
	client, err := pigeon.New(nil)
	if err != nil {
		log.Fatalf("Unable to retrieve vision service: %v\n", err)
	}

	// To call multiple image annotation requests.
	batch, err := client.NewBatchAnnotateImageRequest(detects.Args(), detects.Features()...)
	if err != nil {
		log.Fatalf("Unable to retrieve image request: %v\n", err)
	}

	// initialize a call with cache support, and monthly max limit.
	call := cache.Call(client.ImagesService().Annotate(batch), store, batch).MaxPerMonth(1000)

	// Execute the "vision.images.annotate".
	res, err := call.Do()
	if err != nil {
		log.Fatalf("Unable to execute images annotate requests: %v\n", err)
	}

	// Marshal annotations from responses
	body, err := json.MarshalIndent(res.Responses, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal the response: %v\n", err)
	}
	fmt.Println(string(body))
}
