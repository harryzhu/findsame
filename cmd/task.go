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
	fileCount := 0

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

		fpath = ToUnixSlash(fpath)
		relpath := strings.TrimPrefix(strings.TrimPrefix(fpath, SourceDir), "/")

		fsize = fmt.Sprintf("%v", finfo.Size())
		if fsize == "0" {
			chanEmptyFile <- relpath
		} else {
			//dbUpdatePathSize(relpath, fsize)
			_, err = stmt.Exec(relpath, fsize, fsize)
			PrintError("TaskUpdateFileSize:Exec", err)
		}

		fileCount++
		PrintSpinner(fmt.Sprintf("%d", fileCount))

		return nil
	})
	stmt.Close()
	tx.Commit()

	chanEmptyFile <- doneEmptyEntry

	PrintlnInfo("TaskUpdateFileSize", "Total ", fileCount, ", Elapse ", time.Since(t1))

	return nil
}

func TaskHashFileFromChan() error {
	t1 := time.Now()
	wg := sync.WaitGroup{}
	for {
		ch := <-chanHashFile
		if ch == doneHashEntry {
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

	PrintlnInfo("TaskHashFiles Elapse", time.Since(t1))
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

	chanHashFile <- doneHashEntry

	PrintlnInfo("TaskSelectFiles Elapse", time.Since(t1))
	return nil
}

func TaskExportEmptyFiles() error {
	t1 := time.Now()
	emptyFilePath := filepath.Join(LogDir, "empty-files.html")
	fpempty, err := os.Create(emptyFilePath)
	PrintError("TaskExportEmptyFiles", err)

	var lines []string
	SaveFile(fpempty, styleCSS)
	SaveFile(fpempty, "<body><ol>")
	fcount := 0
	for {
		ch := <-chanEmptyFile
		if ch == doneEmptyEntry {
			break
		}

		if IsCancelAll {
			break
		}

		fcount++
		lines = append(lines, strings.Join([]string{"<li>", ch, "</li>"}, ""))
		if len(lines) > 100 {
			SaveFile(fpempty, strings.Join(lines, ""))
			lines = lines[:0]
		}
	}

	SaveFile(fpempty, strings.Join(lines, ""))
	SaveFile(fpempty, "</ol><hr>Total: "+fmt.Sprintf("%v", fcount)+"</body></html>")
	fpempty.Close()

	PrintlnInfo("TaskExportEmptyFiles Elapse", time.Since(t1))

	return nil
}

func TaskExportSameFiles() error {
	t1 := time.Now()
	//
	sameFilePath := filepath.Join(LogDir, "same-files.html")
	fpsame, err := os.Create(sameFilePath)
	PrintError("TaskExportSameFiles", err)

	var hpsame []string

	sqlCmd := `select fhash from pathash where fhash is NOT NULL and fhash != "" group by fhash having count(*) > 1 order by fpath;`
	rows, err := db.Query(sqlCmd)

	var cfhash string
	for rows.Next() {
		err = rows.Scan(&cfhash)
		PrintError("TaskExportSameFiles", err)
		hpsame = append(hpsame, cfhash)
	}

	//
	SaveFile(fpsame, styleCSS+"<ul>")
	var line string
	var lines []string

	sqlCmd = "select fpath from pathash where fhash = ? order by fpath;"
	stmt, err := db.Prepare(sqlCmd)
	FatalError("TaskExportSameFiles:Prepare", err)

	var cfpaths []string
	var fileCount, duplicateCount int
	for _, cfhash := range hpsame {
		DebugInfo("cfhash", cfhash)
		cfpaths = cfpaths[:0]
		cfpaths = dbGetPathByHash(cfhash, stmt)
		fileCount++
		duplicateCount += len(cfpaths)
		line = strings.Join(cfpaths, "<br>")
		lines = append(lines, strings.Join([]string{"<li>", line, "</li>"}, ""))
		if len(lines) > 100 {
			SaveFile(fpsame, strings.Join(lines, ""))
			lines = lines[:0]
		}
	}

	stmt.Close()

	SaveFile(fpsame, strings.Join(lines, ""))
	SaveFile(fpsame, "</ul><hr>"+fmt.Sprintf("Files: %d, Same(Total): %d", fileCount, duplicateCount)+"<br></body></html>")

	fpsame.Close()

	PrintlnInfo("TaskExportSameFiles Elapse", time.Since(t1))

	return nil
}

func TaskCancelAll() error {
	IsCancelAll = true
	PrintlnInfo(Cyan("TaskCancelAll"), " cancel all ...")
	maxSleep := 0
	for {
		if IsReadyForExit {
			break
		}
		if maxSleep > 10 {
			break
		}
		DebugInfo("TaskCancelAll", "waiting for IsReadyForExit")
		DebugInfo("chan", "chanEmptyFile: ", len(chanEmptyFile), ", chanHashFile: ", len(chanHashFile))
		time.Sleep(time.Second)
		maxSleep++
	}
	return nil
}
