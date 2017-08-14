package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	vision "google.golang.org/api/vision/v1"
)

// Storer of *vision.BatchAnnotateImagesResponse
type Storer interface {
	Init() error
	LastMonthCount() (int, error)
	Get(hash string) (*vision.BatchAnnotateImagesResponse, bool)
	Save(hash string, res *vision.BatchAnnotateImagesResponse) error
}

// FsStore stores on disk
type FsStore struct {
	Path   string
	Expire time.Duration
}

// Init the store directory.
func (f FsStore) Init() error {
	return os.MkdirAll(f.Path, os.ModePerm)
}

// Save a response.
func (f FsStore) Save(hash string, res *vision.BatchAnnotateImagesResponse) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	file := filepath.Join(f.Path, hash+".json")
	return ioutil.WriteFile(file, data, os.ModePerm)
}

// Get a response
func (f FsStore) Get(hash string) (*vision.BatchAnnotateImagesResponse, bool) {
	file := filepath.Join(f.Path, hash+".json")
	stats, err := os.Stat(file)
	if os.IsNotExist(err) {
		return nil, false
	}
	if f.Expire > 0 && time.Now().After(stats.ModTime().Add(f.Expire)) {
		return nil, false
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, false
	}
	v := &vision.BatchAnnotateImagesResponse{}
	return v, json.Unmarshal(data, v) == nil
}

// LastMonthCount count the numbers of hash created in lat 30 days
func (f FsStore) LastMonthCount() (int, error) {
	return 0, fmt.Errorf("not implemented")
}
