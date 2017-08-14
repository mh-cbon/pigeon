package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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

	store := cache.FsStore{Path: filepath.Join(home, ".gvision-cache"), Expire: time.Minute * 30}
	if err := store.Init(); err != nil {
		log.Fatalf("Failed to intialize the cache store: %v\n", err)
	}

	// Initialize vision service by a credentials json.
	client, err := cache.New(store, nil)
	if err != nil {
		log.Fatalf("Unable to retrieve vision service: %v\n", err)
	}

	// To call multiple image annotation requests.
	batch, err := client.NewBatchAnnotateImageRequest(detects.Args(), detects.Features()...)
	if err != nil {
		log.Fatalf("Unable to retrieve image request: %v\n", err)
	}
	// Execute the "vision.images.annotate".
	res, err := client.ImagesService().Annotate(batch).Do()
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
