package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ExportEmptyFiles() error {
	t1 := time.Now()
	emptyFilePath := filepath.Join(LogDir, "empty-files.txt")
	fpempty, err := os.Create(emptyFilePath)
	PrintError("ExportEmptyFiles", err)

	writer := bufio.NewWriter(fpempty)
	for {
		ch := <-chanEmptyFile
		if ch == flagAllDone {
			break
		}

		if IsCancelAll {
			break
		}

		_, err = writer.WriteString(ch + "\n")
		PrintError("ExportEmptyFiles", err)
	}

	writer.Flush()
	fpempty.Close()

	PrintlnInfo("ExportEmptyFiles", time.Since(t1))

	return nil
}

func ExportSameFiles() error {
	t1 := time.Now()
	//
	sameFilePath := filepath.Join(LogDir, "same-files.html")
	fpsame, err := os.Create(sameFilePath)
	PrintError("ExportSameFiles", err)

	var hpsame []string

	sqlCmd := `select fhash from pathash where fhash is NOT NULL and fhash != "" group by fhash having count(*) > 1 order by fpath;`
	rows, err := db.Query(sqlCmd)

	var cfhash string
	for rows.Next() {
		err = rows.Scan(&cfhash)
		PrintError("ExportSameFiles", err)
		hpsame = append(hpsame, cfhash)
	}

	//
	SaveFile(fpsame, styleCSS+"<ul>")
	var line string
	var lines []string

	sqlCmd = "select fpath from pathash where fhash = ? order by fpath;"
	stmt, err := db.Prepare(sqlCmd)
	FatalError("ExportSameFiles:Prepare", err)

	var cfpaths []string
	var fileCount, duplicateCount int
	for _, cfhash := range hpsame {
		DebugInfo("cfhash", cfhash)
		cfpaths = cfpaths[:0]
		cfpaths = dbGetPathByHash(cfhash, stmt)
		fileCount++
		duplicateCount += len(cfpaths)
		line = ""
		for _, cfp := range cfpaths {
			line = strings.Join([]string{cfp, ` <a href="file://`, SourceDir, "/", cfp, `">&hellip;</a><br>`, line}, "")
		}

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

	PrintlnInfo("ExportSameFiles", time.Since(t1))

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
		if maxSleep > 5 {
			break
		}
		DebugInfo("TaskCancelAll", "waiting for IsReadyForExit")
		DebugInfo("chan", "chanHashFile: ", len(chanHashFile))
		time.Sleep(time.Second)
		maxSleep++
	}
	return nil
}
