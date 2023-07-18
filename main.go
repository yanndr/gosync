package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var Version = "0.1.dev"

const bufferSize = 10
const numberOfWorker = 10

var deleteChan = make(chan string, bufferSize)
var copyChan = make(chan string)

func main() {
	var source, destination string

	flag.StringVar(&source, "s", "", "The source folder to sync")
	flag.StringVar(&destination, "d", "", "The destination folder to sync")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Sync v%s is a CLI to synchronize two folders", Version)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	//err := synchronize(source, destination)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(-1)
	//}

	now := time.Now()

	defer func() {
		fmt.Println(time.Since(now))
	}()

	ds := DirectorySync{source: source, destination: destination}
	err := ds.Sync(40)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
