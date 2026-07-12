/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	IsDebug   bool
	IsSerial  bool
	SourceDir string
	LogDir    string

	//
	numCPU    int
	timeStart time.Time
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "findsame",
	Short: "--source-dir=/path/of/folder --log-dir=./logs --debug=true|false --serial=true|false",
	Long:  `a quick tool for finding duplicate files`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		DebugInfo("findsame", "Thanks for choosing findsame!")
		numCPU = runtime.NumCPU()
		if numCPU < 4 {
			numCPU = 4
		}
		SourceDir = ToUnixSlash(SourceDir)
		LogDir = ToUnixSlash(LogDir)

		MakeDirs(LogDir)

		timeStart = time.Now()
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		_, err := os.Stat(SourceDir)
		if err != nil {
			FatalError(strings.Join([]string{"--source-dir=", SourceDir}, ""), err)
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		dbInit()

		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			ExportEmptyFiles()
		}()
		go func() {
			defer wg.Done()
			TaskUpdateFileSize()
		}()
		wg.Wait()

		//
		dbUpdateHashBySameSize()

		//
		ExportSameFiles()

		dbClose()

		IsReadyForExit = true
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		PrintlnInfo("save result into", LogDir)
		PrintlnInfo("\n\n::: Total Elapse", time.Since(timeStart), " :::\n")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&IsDebug, "debug", false, "if print debug info")
	rootCmd.PersistentFlags().BoolVar(&IsSerial, "serial", false, "if you are using HDD(not SSD), pls set --serial=true")
	//
	rootCmd.PersistentFlags().StringVar(&SourceDir, "source-dir", "", "root dir for finding same files")
	rootCmd.PersistentFlags().StringVar(&LogDir, "log-dir", "./logs", "path for saving result")
}
