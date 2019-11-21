package main

import (
	"flag"
	"fmt"
	. "github.com/liserjrqlxue/simple-util"
	"log"
	"os"
	"strings"
	"time"
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
	lines := make(chan string)
	go func(pipeline string, lines chan string) {
		runPipeline(pipeline, lines)
		close(lines)
	}(*pipeline, lines)
	createGraphviz(*pipeline, lines)
	for i := 0; i < *threshold; i++ {
		throttle <- true
	}
}

func runPipeline(pipeline string, lines chan<- string) {
	//log.Println("runPipeline:" + pipeline)
	var suffix = getSuffix(pipeline)
	switch suffix {
	case "sh":
		runScript(pipeline, lines)
	case "step":
		runStep(pipeline, lines)
	case "parallel":
		runParallel(pipeline, lines)
	}
}

func getSuffix(str string) string {
	array := strings.Split(str, ".")
	//log.Println("getSuffix(" + str + ")->" + array[len(array)-1])
	return array[len(array)-1]
}

func runScript(script string, lines chan<- string) {
	throttle <- true
	log.Println("runScript:" + script)
	time.Sleep(1 * time.Second)
	CheckErr(RunCmd("bash", script), "run "+script+" error!")
	log.Println("runScript:" + script + "\tDone")
	lines <- "\"" + script + "\""
	<-throttle
}

func runStep(step string, lines chan<- string) {
	log.Println("runStep:" + step)
	array := File2Array(step)
	lines <- "\"" + step + "\"->\"" + strings.Join(array, "\"->\"") + "\""
	for _, item := range array {
		runPipeline(item, lines)
	}
}

func runParallel(parallel string, lines chan<- string) {
	log.Println("runParallel:" + parallel)
	array := File2Array(parallel)
	chanList := make(chan int, len(array))
	for i, item := range array {
		go func(i int, item string) { // parallel
			lines <- "\"" + parallel + "\"->\"" + item + "\""
			runPipeline(item, lines)
			chanList <- i
		}(i, item)
	}
	for range array {
		<-chanList
	}
}

func createGraphviz(prefix string, lines <-chan string) {
	dot, err := os.Create(prefix + ".dot")
	CheckErr(err)
	defer DeferClose(dot)
	_, err = fmt.Fprintln(dot, "digraph G {")
	CheckErr(err)
	for item := range lines {
		_, err = fmt.Fprintln(dot, item)
		CheckErr(err)
	}
	_, err = fmt.Fprintln(dot, "}")
	CheckErr(err)
}
