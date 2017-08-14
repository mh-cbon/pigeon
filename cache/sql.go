package cache

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	vision "google.golang.org/api/vision/v1"
)

// Sqlite panics if it cant initialize and open a db in given path.
func Sqlite(path string) *sql.DB {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		panic(err)
	}
	file := filepath.Join(path, "db.sqlite")
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		panic(err)
	}
	return db
}

// SQLStore stores in sqlite db
type SQLStore struct {
	Expire time.Duration
	Db     *sql.DB
}

// Init the sqlite db.
func (s *SQLStore) Init() error {
	sql := `
	CREATE TABLE IF NOT EXISTS items(
		hash TEXT,
		data TEXT,
		createdate DATETIME
	);
	`
	_, err := s.Db.Exec(sql)
	return err
}

// Save a response.
func (s *SQLStore) Save(hash string, res *vision.BatchAnnotateImagesResponse) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	sql := `
	INSERT INTO items(
		hash,
		data,
		createdate
	) values(?, ?, CURRENT_TIMESTAMP)
	`
	stmt, err := s.Db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err2 := stmt.Exec(hash, string(data))
	return err2
}

// Get a response
func (s *SQLStore) Get(hash string) (*vision.BatchAnnotateImagesResponse, bool) {

	sql := `
		SELECT data, createdate FROM items WHERE hash= ? LIMIT 1
		`
	stmt, err := s.Db.Prepare(sql)
	if err != nil {
		return nil, false
	}
	defer stmt.Close()

	rows, err := stmt.Query(hash)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var data string
	var createdate time.Time
	rows.Next()
	if err := rows.Scan(&data, &createdate); err != nil {
		return nil, false
	}
	if s.Expire > 0 && time.Now().After(createdate.Add(s.Expire)) {
		return nil, false
	}
	v := &vision.BatchAnnotateImagesResponse{}
	return v, json.Unmarshal([]byte(data), v) == nil
}

// LastMonthCount count the numbers of hash created in lat 30 days
func (s *SQLStore) LastMonthCount() (int, error) {
	sql := `
		SELECT COUNT(*) FROM items WHERE createdate > date('now','-1 month')
		`
	stmt, err := s.Db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var c int
	rows.Next()
	if err := rows.Scan(&c); err != nil {
		return 0, err
	}
	return c, nil
}
