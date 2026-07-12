package cmd

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

const (
	minBlockHashSize int = 8 << 20
)

func dbInit() error {
	openDB()
	err := db.Ping()
	FatalError("dbInit", err)

	sqlCmd := `CREATE TABLE IF NOT EXISTS pathash(id INTEGER PRIMARY KEY AUTOINCREMENT, fpath TEXT, fsize TEXT, bhash TEXT, fhash TEXT);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_fpath ON pathash(fpath);`

	DebugInfo("dbInit", sqlCmd)

	_, err = db.Exec(sqlCmd)
	FatalError("dbInit", err)

	return nil
}

func dbGetPathByHash(fhash string, stmt *sql.Stmt) (files []string) {
	rows, err := stmt.Query(fhash)
	if err != nil {
		PrintError("dbGetPathByHash: stmt.Query", err)
		return files
	}

	var fpath string
	for rows.Next() {
		err = rows.Scan(&fpath)
		if err != nil {
			PrintError("dbGetPathByHash", err)
			continue
		}
		files = append(files, fpath)
	}
	rows.Close()

	DebugInfo("dbGetPathByHash", fhash, " :: ", files)
	return files
}

func dbDeleteUniqueBySize() error {
	t1 := time.Now()
	//
	tx, err := db.Begin()
	FatalError("dbDeleteUniqueBySize:Begin", err)
	sqlCmd := `DELETE FROM pathash WHERE fsize IN 
	(select fsize from pathash group by fsize having count(*) = 1);`
	stmt, err := tx.Prepare(sqlCmd)
	FatalError("dbDeleteUniqueBySize:Prepare", err)

	res, err := stmt.Exec()
	PrintError("dbDeleteUniqueBySize", err)
	deleteCount, err := res.RowsAffected()
	PrintError("dbDeleteUniqueBySize", err)

	stmt.Close()
	tx.Commit()

	DebugInfo("dbDeleteUniqueBySize", fmt.Sprintf("%d", deleteCount), ", Elapse ", time.Since(t1))

	return nil
}

func dbUpdateHashBySameSize() error {
	t1 := time.Now()
	//
	dbDeleteUniqueBySize()
	//
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		TaskHashFileFromChan()
	}()

	go func() {
		defer wg.Done()
		TaskSelectFilesForHash()
	}()

	wg.Wait()

	//
	tx, err := db.Begin()
	FatalError("dbUpdateHashBySameSize:Begin", err)
	sqlCmd := "INSERT INTO pathash(fpath,fhash) VALUES(?,?) ON CONFLICT(fpath) DO UPDATE SET fhash = ?;"
	stmt, err := tx.Prepare(sqlCmd)
	FatalError("dbUpdateHashBySameSize:Prepare", err)

	hashTotal := 0
	safePathHash.Range(func(key, val any) bool {
		k := key.(string)
		v := val.(string)
		if k != "" && v != "" {
			hashTotal++
			stmt.Exec(k, v, v)
		}
		return true
	})

	stmt.Close()
	tx.Commit()

	PrintlnInfo("dbUpdateHashBySameSize", "Hash ", hashTotal, ", Elapse ", time.Since(t1))
	return nil
}

func dbClose() error {
	err := db.Close()
	DebugInfo("dbClose", "closing db")
	PrintError("dbClose", err)
	return nil
}
