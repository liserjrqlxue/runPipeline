package main

import (
	"flag"
	"fmt"
	. "github.com/liserjrqlxue/simple-util"
	"log"
	"os"
	"strconv"
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
		log.Println("Start:\t" + pipeline)
		runPipeline(pipeline, "", lines, 0)
		close(lines)
	}(*pipeline, lines)
	createGraphviz(*pipeline, lines)
	for i := 0; i < *threshold; i++ {
		throttle <- true
	}
	log.Println("End:\t" + *pipeline)
}

func runPipeline(pipeline, flag string, lines chan<- string, rank int) {
	//log.Println("runPipeline:" + pipeline)
	var suffix = getSuffix(pipeline)
	switch suffix {
	case "sh":
		runScript(pipeline, flag, lines, rank)
	case "step":
		runStep(pipeline, flag, lines, rank)
	case "parallel":
		runParallel(pipeline, flag, lines, rank)
	}
}

func getSuffix(str string) string {
	array := strings.Split(str, ".")
	//log.Println("getSuffix(" + str + ")->" + array[len(array)-1])
	return array[len(array)-1]
}

func runScript(script, flag string, lines chan<- string, rank int) {
	throttle <- true
	log.Printf("%s%d:\t%s", flag, rank, script)
	time.Sleep(1 * time.Second)
	//CheckErr(RunCmd("bash", script), "run "+script+" error!")
	log.Printf("%s%d:\t%s\tDone", flag, rank, script)
	lines <- "\"" + script + "\""
	<-throttle
}

func runStep(step, flag string, lines chan<- string, rank int) {
	if flag == "" && rank == 0 {
		flag = "step"
		log.Printf("%s:\t%s", flag, step)
	} else {
		flag += strconv.Itoa(rank)
		log.Printf("%s:\t%s", flag, step)
		flag += ".step"
	}

	array := File2Array(step)
	lines <- "\"" + step + "\"->\"" + strings.Join(array, "\"->\"") + "\""
	for i, item := range array {
		runPipeline(item, flag, lines, i)
	}
	log.Printf("%s:\t%s\tDone", flag, step)
}

func runParallel(parallel, flag string, lines chan<- string, rank int) {
	if flag == "" && rank == 0 {
		flag = "para"
		log.Printf("%s:\t%s", flag, parallel)
	} else {
		flag += strconv.Itoa(rank)
		log.Printf("%s:\t%s", flag, parallel)
		flag += ".para"
	}
	array := File2Array(parallel)
	chanList := make(chan int, len(array))
	for i, item := range array {
		go func(i int, item string) { // parallel
			lines <- "\"" + parallel + "\"->\"" + item + "\""
			runPipeline(item, flag, lines, i)
			chanList <- i
		}(i, item)
		rank++
	}
	for range array {
		<-chanList
	}
	log.Printf("%s:\t%s\tDone", flag, parallel)
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
