package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func TaskUpdateFileSize() error {
	t1 := time.Now()
	var fsize string
	var fileCount int = 0

	tx, err := db.Begin()
	FatalError("TaskUpdateFileSize:Begin", err)
	sqlCmd := strings.Join([]string{"INSERT INTO pathash(fpath,fsize) VALUES(?,?) ON CONFLICT(fpath) DO UPDATE SET fsize = ?;"}, "")
	stmt, err := tx.Prepare(sqlCmd)
	FatalError("TaskUpdateFileSize:Prepare", err)
	filepath.Walk(SourceDir, func(fpath string, finfo os.FileInfo, err error) error {
		if IsCancelAll == true {
			return nil
		}

		if err != nil {
			PrintError("TaskUpdateFileSize", err)
			return err
		}

		if finfo.IsDir() {
			return nil
		}

		fileCount++
		PrintSpinner(fmt.Sprintf("%d,em: %d", fileCount, len(chanEmptyFile)))

		fpath = ToUnixSlash(fpath)
		relpath := strings.TrimPrefix(strings.TrimPrefix(fpath, SourceDir), "/")

		if finfo.Size() == 0 {
			chanEmptyFile <- relpath
			return nil
		}

		fsize = fmt.Sprintf("%v", finfo.Size())
		_, err = stmt.Exec(relpath, fsize, fsize)
		PrintError("TaskUpdateFileSize:Exec", err)

		return nil
	})
	chanEmptyFile <- flagAllDone

	stmt.Close()
	tx.Commit()

	PrintlnInfo("TaskUpdateFileSize", "Total ", fileCount, ", Elapse ", time.Since(t1))

	return nil
}

func TaskHashFileFromChan() error {
	t1 := time.Now()
	wg := sync.WaitGroup{}
	for {
		ch := <-chanHashFile
		if ch == flagAllDone {
			break
		}

		if IsCancelAll {
			break
		}

		wg.Add(1)
		go func(ch string, SourceDir string) {
			defer wg.Done()
			safePathHash.Store(ch, GetXxhashFile(ToUnixSlash(filepath.Join(SourceDir, ch))))
		}(ch, SourceDir)

		if IsSerial {
			wg.Wait()
		}
	}
	wg.Wait()

	PrintlnInfo("TaskHashFiles", time.Since(t1))
	return nil
}

func TaskSelectFilesForHash() error {
	t1 := time.Now()
	sqlCmd := `select fpath from pathash order by fpath;`
	rows, err := db.Query(sqlCmd)

	var fpath string
	var processCount int = 0
	for rows.Next() {
		err = rows.Scan(&fpath)
		PrintError("TaskSelectFiles", err)
		chanHashFile <- fpath

		processCount++
		PrintSpinner(fmt.Sprintf("%d +++", processCount))
	}
	rows.Close()

	chanHashFile <- flagAllDone

	PrintlnInfo("TaskSelectFiles", time.Since(t1))
	return nil
}
