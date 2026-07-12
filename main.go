/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"findsame/cmd"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(3)

	var IsMainReadyForExit bool = false
	go func() {
		cmd.Execute()
		IsMainReadyForExit = true
	}()

	go func() {
		onExit()
	}()

	go func() {
		for {
			if IsMainReadyForExit {
				os.Exit(0)
			}
		}
	}()

	wg.Wait()
}

func onExit() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cmd.TaskCancelAll()
		time.Sleep(time.Second)
		os.Exit(0)
	}()
}
