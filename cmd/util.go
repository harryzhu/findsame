package cmd

import (
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/cespare/xxhash/v2"
)

const (
	bufSize int = 1024 * 128
)

func MakeDirs(dpath string) error {
	_, err := os.Stat(dpath)
	if err != nil {
		DebugInfo("MakeDirs", dpath)
		err = os.MkdirAll(dpath, os.ModePerm)
		PrintError("MakeDirs:MkdirAll", err)
	}
	return nil
}

func ToUnixSlash(s string) string {
	// for windows
	return strings.ReplaceAll(s, "\\", "/")
}

func GetXxhashFile(fpath string) string {
	fin, err := os.Open(fpath)
	PrintError("GetXxhashFile", err)
	defer fin.Close()

	hasher := xxhash.New()
	buffer := make([]byte, bufSize)
	if _, err := io.CopyBuffer(hasher, fin, buffer); err != nil {
		PrintError("GetXxhashFile", err)
		return ""
	}

	return strconv.FormatUint(hasher.Sum64(), 10)
}

func SaveFile(fp *os.File, fdata string) error {
	_, err := fp.WriteString(fdata)
	FatalError("SaveFile", err)
	return nil
}

func RemoveFile(fpath string) error {
	if fpath == "" {
		return nil
	}
	_, err := os.Stat(fpath)
	if err != nil {
		return nil
	}
	return os.Remove(fpath)
}
