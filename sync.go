package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	sync2 "sync"
)

func isAValidDirectory(path string) error {
	sourceInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("%s is not a valid directory", path)
	}

	return nil
}

func sync(source, destination string) error {

	var fileToSync = make(map[string]bool)
	var directoryToDelete = make(map[string]bool)

	if err := isAValidDirectory(source); err != nil {
		return err
	}

	if err := isAValidDirectory(destination); err != nil {
		return err
	}

	err := walk(source, "", func(filePath string) {
		fmt.Println(filePath)
		fileToSync[filePath] = true
	}, func(path string) {
		fmt.Println(path)
		directoryToDelete[path] = true
	})

	if err != nil {
		return err
	}

	errorC := make(chan error)
	go func() {
		for err := range errorC {
			fmt.Println(err)
		}
	}()
	deleteDoneC := workers(numberOfWorker, deleteChan, errorC, func(path string) error {
		err := os.Remove(path)
		if err != nil {
			return err
		}
		return nil
	})

	err = walk(destination, "", func(filePath string) {
		fmt.Println(filePath)
		if fileToSync[filePath] {
			delete(fileToSync, filePath)
		} else {
			deleteChan <- path.Join(destination, filePath)
		}
	}, func(directoryPath string) {
		if !directoryToDelete[directoryPath] {
			deleteChan <- path.Join(destination, directoryPath)
		}
	})

	close(deleteChan)

	if err != nil {
		return err
	}

	copyDoneC := workers(numberOfWorker, copyChan, errorC, copyFile(source, destination))

	for k := range fileToSync {
		copyChan <- k
	}
	close(copyChan)

	<-deleteDoneC
	<-copyDoneC

	return nil
}

func walk(dir, subDir string, fileFn func(path string), directoryFn func(path string)) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			directoryFn(path.Join(subDir, e.Name()))
			err = walk(path.Join(dir, e.Name()), e.Name(), fileFn, directoryFn)
		} else {
			fileFn(path.Join(subDir, e.Name()))
		}
	}

	return nil
}

func deleteWorkers(errorC chan<- error) <-chan bool {
	doneC := make(chan bool, 0)
	for i := 0; i < numberOfWorker; i++ {
		go func() {
			for file := range deleteChan {
				err := os.RemoveAll(file)
				if err != nil {
					errorC <- err
				}
			}

			doneC <- true
		}()
	}

	return doneC
}

func copyWorkers(errorC chan<- error) <-chan bool {
	doneC := make(chan bool, 0)
	for i := 0; i < numberOfWorker; i++ {
		go func() {
			for file := range deleteChan {
				err := os.Remove(file)
				if err != nil {
					errorC <- err
				}
			}

			doneC <- true
		}()
	}

	return doneC
}

func workers(numberOfWorker int, inputChan <-chan string, errorC chan<- error, fn func(path string) error) <-chan bool {
	doneC := make(chan bool, 0)
	wg := sync2.WaitGroup{}
	for i := 0; i < numberOfWorker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range inputChan {
				err := fn(p)
				if err != nil {
					errorC <- err
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		doneC <- true
	}()

	return doneC
}

func copyFile(source, destination string) func(filePath string) error {
	return func(filePath string) error {
		fmt.Printf("copy %s %s to %s\n", filePath, source, destination)
		src, err := os.Open(path.Join(source, filePath))
		if err != nil {
			return err
		}
		defer src.Close()

		destinationFile := path.Join(destination, filePath)
		dir := filepath.Dir(destinationFile)
		_, err = os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					return err
				}
			} else {
				return err
			}

		}

		destination, err := os.Create(destinationFile)
		if err != nil {
			return err
		}
		defer destination.Close()
		_, err = io.Copy(destination, src)
		if err != nil {
			return err
		}

		return nil
	}
}
