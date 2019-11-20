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
	threshold = flag.Int(
		"threshold",
		12,
		"threshold to used",
	)
)

var throttle = make(chan bool)

func main() {
	flag.Parse()
	if *pipeline == "" {
		flag.Usage()
		log.Fatal("-pipeline is required!")
	}
	throttle = make(chan bool, *threshold)
	runPipeline(*pipeline)
	for i := 0; i < *threshold; i++ {
		throttle <- true
	}
}

func runPipeline(pipeline string) {
	log.Println("runPipeline:" + pipeline)
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
	log.Println("getSuffix(" + str + ")->" + array[len(array)-1])
	return array[len(array)-1]
}

func runScript(script string) {
	throttle <- true
	log.Println("runScript:" + script)
	CheckErr(RunCmd("bash", script), "run "+script+" error!")
	<-throttle
}

func runStep(step string) {
	log.Println("runStep:" + step)
	array := File2Array(step)
	for _, item := range array {
		runPipeline(item)
	}
}

func runParallel(parallel string) {
	log.Println("runParallel:" + parallel)
	array := File2Array(parallel)
	chanList := make(chan int, len(array))
	for i, item := range array {
		go func(i int, item string) { // parallel
			runPipeline(item)
			chanList <- i
		}(i, item)
	}
	for range array {
		<-chanList
	}
}
