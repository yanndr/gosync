package main

import (
	"errors"
	"flag"
	"fmt"
	"gosync/pkg/directory"
	"os"
)

var Version = "0.1.dev"

func main() {
	var source, destination string

	flag.StringVar(&source, "s", "", "The source folder to synchronize")
	flag.StringVar(&destination, "d", "", "The destination folder to synchronize")

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

	ds := directory.NewSynchronizer(source, destination, directory.MaxGoroutine(40))

	err := ds.Sync()
	if err != nil {

		var cpErr *directory.CopyError
		if errors.As(err, &cpErr) {
			fmt.Printf("Process ended with errors:\n%s\n", cpErr.Error())
			os.Exit(1)
		}

		fmt.Println(err)

		var inputErr *directory.InputError
		if errors.As(err, &inputErr) {
			os.Exit(2)
		}

		os.Exit(255)
	}
}
