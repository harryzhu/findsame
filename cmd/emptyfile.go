package cmd

import (
	"bufio"
	"os"
	"path/filepath"
	"time"
)

func TaskExportEmptyFiles() error {
	t1 := time.Now()
	emptyFilePath := filepath.Join(LogDir, "empty-files.txt")
	fpempty, err := os.Create(emptyFilePath)
	PrintError("TaskExportEmptyFiles", err)

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
		PrintError("TaskExportEmptyFiles", err)
	}

	writer.Flush()
	fpempty.Close()

	PrintlnInfo("TaskExportEmptyFiles Elapse", time.Since(t1))

	return nil
}
