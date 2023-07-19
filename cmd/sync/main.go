package main

import (
	"flag"
	"fmt"
	"gosync/pkg/directory"
	"os"
	"time"
)

var Version = "0.1.dev"

func main() {
	var source, destination string

	flag.StringVar(&source, "s", "", "The source folder to synchronizeFolder")
	flag.StringVar(&destination, "d", "", "The destination folder to synchronizeFolder")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Sync v%s is a CLI to synchronize two folders", Version)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if source == "" || destination == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	now := time.Now()

	defer func() {
		fmt.Println(time.Since(now))
	}()

	ds := directory.NewSynchronizer(source, destination, directory.MaxGoroutine(40))

	err := ds.Sync()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
