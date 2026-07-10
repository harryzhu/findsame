//go:build (darwin && arm64) || (darwin && amd64)
// +build darwin,arm64 darwin,amd64

package cmd

import (
	"database/sql"
	"path/filepath"

	// CGO needed
	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

func openDB() error {
	dbpath := filepath.Join(LogDir, "data")

	err := RemoveFile(dbpath)
	FatalError("openDB", err)
	RemoveFile(filepath.Join(LogDir, "empty-files.html"))
	RemoveFile(filepath.Join(LogDir, "same-files.html"))

	db, err = sql.Open("sqlite3", dbpath)
	FatalError("openDB", err)

	return nil
}
