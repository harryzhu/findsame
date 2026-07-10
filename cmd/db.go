//go:build (linux && amd64) || (windows && amd64)
// +build linux,amd64 windows,amd64

package cmd

import (
	"database/sql"
	"path/filepath"

	// no  CGO needed
	//_ "github.com/glebarez/go-sqlite"
	_ "modernc.org/sqlite"
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

	db, err = sql.Open("sqlite", dbpath)
	FatalError("openDB", err)

	return nil
}
