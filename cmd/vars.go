package cmd

import (
	"sync"
)

func init() {
	flagAllDone = "__ALLDONE__"
}

var chanHashFile chan string = make(chan string, numCPU*2)
var chanHashBlock chan string = make(chan string, numCPU*2)

var chanEmptyFile chan string = make(chan string, 4096)

var flagAllDone string

var safePathHash sync.Map

var IsCancelAll bool = false

var IsReadyForExit bool = false

var styleCSS string = `<!doctype html><meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
	<meta name="keywords" content="请在Chrome中打开" />
	<title>Same Files|重复文件</title>
	<style>
	body{
		margin-left: auto;
		margin-right: auto;
		max-width: 960px;
		}
	a{
		text-decoration: none;
		color: #ccc;
		}
	hr{
		border: none;
		border-bottom: 1px solid #ccc;
		}
	ul,li{
		list-style:none;
		padding: 0;
		margin:0;
		}
	li{
		border: 1px solid #ccc;
		border-radius: 6px;
		margin-top: 10px;
		padding: 10px;
		background-color: #f2f2f2;
		}
	span.cfhash{
		padding: 2px 5px;
		background-color: #ccc;
		}
	</style>
	`
