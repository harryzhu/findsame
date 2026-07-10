package cmd

import (
	"sync"
)

func init() {
	doneSameEntry["__ALLDONE__"] = "__ALLDONE__"
	doneEmptyEntry = "__ALLDONE__"
}

var chanPathHash chan map[string]string = make(chan map[string]string, 8192)
var doneSameEntry map[string]string = make(map[string]string, 1)

var chanEmptyFile chan string = make(chan string, 8192)
var doneEmptyEntry string

var IsCancelAll bool = false

var IsReadyForExit bool = false

var styleCSS string = `<!doctype html><meta charset="utf-8">
	<style>
	body{
		margin-left: auto;
		margin-right: auto;
		max-width: 960px;
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

var safePathHash sync.Map
