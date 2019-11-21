package main

import (
	"flag"
	"fmt"
	. "github.com/liserjrqlxue/simple-util"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
	var suffix = getSuffix(pipeline)
	array := File2Array(pipeline)
	if flag == "" && rank == 0 {
		flag = "main"
	} else {
		flag += strconv.Itoa(rank)
	}
	log.Printf("%s\t%s", flag, pipeline)
	switch suffix {
	case "sh":
		throttle <- true
		//time.Sleep(1 * time.Second)
		CheckErr(RunCmd("bash", pipeline), "run "+pipeline+" error!")
		lines <- "\"" + pipeline + "\" [shape=box label=\"" + filepath.Base(pipeline) + "\"]"
		<-throttle
	case "step":
		lines <- "\"" + pipeline + "\" [style=filled color=blue label=\"" + filepath.Base(pipeline) + "\"]"
		lines <- fmt.Sprintf(
			"subgraph cluster_%s{\n"+
				"label=\"%s\"\n"+
				"color=blue\n"+
				"node [style=filled]\n"+
				"\"%s\"[shape=Mdiamond label=Start]\n"+
				"\"%s\"[shape=Msquare label=End]",
			strings.Replace(flag, ".", "_", -1), pipeline, pipeline+".Start", pipeline+".End",
		)
		lines <- "\"" + pipeline + ".Start\"->\"" + strings.Join(array, "\"->\"") + "\"->\"" + pipeline + ".End\""
		lines <- "}"
		lines <- fmt.Sprintf("\"%s\"->\"%s\"", pipeline, pipeline+".Start")
		flag += ".step"
		for i, item := range array {
			runPipeline(item, flag, lines, i)
		}
	case "parallel":
		lines <- "\"" + pipeline + "\" [style=filled color=red label=\"" + filepath.Base(pipeline) + "\"]"
		lines <- fmt.Sprintf(
			"subgraph cluster_%s{\n"+
				"label=\"%s\"\n"+
				"style=filled\n"+
				"color=lightgrey\n"+
				"node [style=filled,color=white]\n"+
				"\"%s\"[shape=Mdiamond label=Start]\n"+
				"\"%s\"[shape=Msquare label=End]",
			strings.Replace(flag, ".", "_", -1), pipeline, pipeline+".Start", pipeline+".End",
		)
		for _, item := range array {
			lines <- fmt.Sprintf("\"%s\"->\"%s\"->\"%s\"", pipeline+".Start", item, pipeline+".End")
		}
		lines <- "}"
		lines <- fmt.Sprintf("\"%s\"->\"%s\"", pipeline, pipeline+".Start")
		flag += ".para"
		chanList := make(chan int, len(array))
		for i, item := range array {
			go func(i int, item string) { // parallel
				runPipeline(item, flag, lines, i)
				chanList <- i
			}(i, item)
			rank++
		}
		for range array {
			<-chanList
		}
	default:
		log.Printf("%s\t%s\tSkip!", flag, pipeline)
		return
	}
	//log.Printf("%s\t%s\tDone", flag, pipeline)
}

func getSuffix(str string) string {
	array := strings.Split(str, ".")
	//log.Println("getSuffix(" + str + ")->" + array[len(array)-1])
	return array[len(array)-1]
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
