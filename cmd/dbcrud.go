package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func dbInit() error {
	openDB()
	err := db.Ping()
	FatalError("dbInit", err)

	sqlCmd := `CREATE TABLE IF NOT EXISTS pathash(id INTEGER PRIMARY KEY AUTOINCREMENT, fpath TEXT, fsize TEXT, fhash TEXT);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_fpath ON pathash(fpath);`

	DebugInfo("dbInit", sqlCmd)

	_, err = db.Exec(sqlCmd)
	FatalError("dbInit", err)

	return nil
}

func dbUpdatePathHash(cfpath, cfhash string) error {
	sqlCmd := strings.Join([]string{`INSERT INTO pathash(fpath,fhash) VALUES("`,
		cfpath, `","`, cfhash, `") ON CONFLICT(fpath) DO UPDATE SET fhash = "`, cfhash, `";`}, "")
	//DebugInfo("UpdatePathHash", sqlCmd)
	_, err := db.Exec(sqlCmd)
	if err != nil {
		PrintError("UpdatePathHash", err)
		return err
	}

	return nil
}

func dbUpdatePathSize(cfpath, cfsize string) error {
	sqlCmd := strings.Join([]string{`INSERT INTO pathash(fpath,fsize) VALUES("`,
		cfpath, `","`, cfsize, `") ON CONFLICT(fpath) DO UPDATE SET fsize = "`, cfsize, `";`}, "")
	//DebugInfo("UpdatePathSize", sqlCmd)
	_, err := db.Exec(sqlCmd)
	if err != nil {
		PrintError("UpdatePathSize", err)
		return err
	}

	return nil
}

func dbDump() (ph map[string]string) {
	ph = make(map[string]string, 8192)
	sqlCmd := "select fpath,fhash from pathash order by fhash;"
	DebugInfo("dbDump", sqlCmd)
	rows, err := db.Query(sqlCmd)
	PrintError("dbDump", err)

	var cfpath string
	var cfhash string
	for rows.Next() {
		err = rows.Scan(&cfpath, &cfhash)
		PrintError("dbDump", err)
		ph[cfpath] = cfhash
	}
	return ph
}

func dbGetPathByHash(fhash string) (files []string) {
	sqlCmd := `select fpath from pathash where fhash ="` + fhash + `" order by fpath;`
	rows, err := db.Query(sqlCmd)

	var fpath string
	for rows.Next() {
		err = rows.Scan(&fpath)
		PrintError("dbGetPathByHash", err)
		files = append(files, fpath)
	}
	rows.Close()
	DebugInfo("dbGetPathByHash", fhash, " :: ", files)
	return files
}

func dbGetPathBySize(fsize string) (files []string) {
	sqlCmd := `select fpath from pathash where fsize ="` + fsize + `" order by fpath;`
	rows, err := db.Query(sqlCmd)

	var fpath string
	for rows.Next() {
		err = rows.Scan(&fpath)
		PrintError("dbGetPathBySize", err)
		files = append(files, fpath)
	}
	rows.Close()
	DebugInfo("dbGetPathBySize", fsize, " :: ", files)
	return files
}

func dbUpdateHashBySameSize() error {
	t1 := time.Now()
	//
	sqlCmd := `select fsize from pathash group by fsize having count(*) > 1;`
	rows, err := db.Query(sqlCmd)

	var fsize string
	var fsizelist []string
	for rows.Next() {
		err = rows.Scan(&fsize)
		PrintError("dbUpdateHashBySameSize", err)
		fsizelist = append(fsizelist, fsize)
	}
	rows.Close()

	if len(fsizelist) > 10 {
		DebugInfo("same size", fsizelist[0:10], " ...")
	} else {
		DebugInfo("same size", fsizelist)
	}

	//
	tx, err := db.Begin()
	FatalError("dbUpdateHashBySameSize:Begin", err)
	sqlCmd = "UPDATE pathash SET fhash = ? where fsize = ?;"
	stmt, err := tx.Prepare(sqlCmd)
	FatalError("dbUpdateHashBySameSize:Prepare", err)
	for _, size := range fsizelist {
		_, err = stmt.Exec("-", size)
		PrintError("dbUpdateHashBySameSize", err)
	}
	stmt.Close()
	tx.Commit()

	//
	sqlCmd = `select fpath from pathash where fhash ="-";`
	rows, err = db.Query(sqlCmd)

	var fpath string
	var processCount int = 0
	wg := sync.WaitGroup{}
	hashCount := int32(0)
	for rows.Next() {
		err = rows.Scan(&fpath)
		PrintError("dbUpdateHashBySameSize", err)

		wg.Add(1)
		go func(fpath string) {
			defer wg.Done()
			atomic.AddInt32(&hashCount, 1)
			safePathHash.Store(fpath, GetXxhashFile(ToUnixSlash(filepath.Join(SourceDir, fpath))))
			atomic.AddInt32(&hashCount, -1)
		}(fpath)

		if IsSerial {
			wg.Wait()
		} else {
			curHashCount := atomic.LoadInt32(&hashCount)
			if curHashCount > int32(numCPU*2) {
				wg.Wait()
			}
		}

		processCount++
		PrintSpinner(fmt.Sprintf("%d +++", processCount))

	}
	wg.Wait()
	rows.Close()

	//
	hashTotal := 0
	safePathHash.Range(func(key, val any) bool {
		k := key.(string)
		v := val.(string)
		if k != "" && v != "" {
			hashTotal++
			dbUpdatePathHash(k, v)
		}
		return true
	})

	chanPathHash <- doneSameEntry

	PrintlnInfo("dbUpdateHashBySameSize", "Hash ", hashTotal, ", Elapse ", time.Since(t1))
	return nil
}

func dbClose() error {
	err := db.Close()
	DebugInfo("dbClose", "closing db")
	PrintError("dbClose", err)
	return nil
}
