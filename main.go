package main

import (
	"flag"
	. "github.com/liserjrqlxue/simple-util"
	"log"
	"strings"
)

var (
	pipeline = flag.String(
		"pipeline",
		"",
		"main pipeline to run",
	)
)

func main() {
	flag.Parse()
	if *pipeline == "" {
		flag.Usage()
		log.Fatal("-pipeline is required!")
	}
	runPipeline(*pipeline)
}

func runPipeline(pipeline string) {
	var suffix = getSuffix(pipeline)
	switch suffix {
	case "sh":
		runScript(pipeline)
	case "step":
		runStep(pipeline)
	case "parallel":
		runParallel(pipeline)
	}
}

func getSuffix(str string) string {
	array := strings.Split(str, ".")
	return array[len(array)-1]
}

func runScript(script string) {
	CheckErr(RunCmd("bash", script), "run "+script+" error!")
}

func runStep(step string) {
	array := File2Array(step)
	for _, item := range array {
		runPipeline(item)
	}
}

func runParallel(parallel string) {
	array := File2Array(parallel)
	chanList := make(chan int, len(array))
	for i, item := range array {
		go func(i int, item string) { // parallel
			runParallel(item)
			chanList <- i
		}(i, item)
	}
	for range array {
		<-chanList
	}
}
