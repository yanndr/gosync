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
	deleteDoneC := deleteWorkers(errorC)

	err = walk(destination, "", func(filePath string) {
		fmt.Println(filePath)
		if fileToSync[filePath] {
			delete(fileToSync, filePath)
		} else {
			deleteChan <- path.Join(destination, filePath)
		}
	}, func(directoryPath string) {
		if directoryToDelete[directoryPath] {
			//deleteChan <- path.Join(destination, directoryPath)
			delete(directoryToDelete, directoryPath)
		}
	})

	close(deleteChan)

	if err != nil {
		return err
	}

	var wg2 sync2.WaitGroup
	for i := 0; i < numberOfWorker; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			for file := range copyChan {
				src, err := os.Open(path.Join(source, file))
				if err != nil {
					fmt.Println(err)
				}
				defer src.Close()

				destinationFile := path.Join(destination, file)
				dir := filepath.Dir(destinationFile)
				_, err = os.Stat(dir)
				if os.IsNotExist(err) {
					os.MkdirAll(dir, os.ModePerm)
				}
				destination, err := os.Create(destinationFile)
				if err != nil {
					fmt.Println(err)
				}
				defer destination.Close()
				_, err = io.Copy(destination, src)
				if err != nil {
					fmt.Println(err)
				}
			}
		}()
	}

	for k := range fileToSync {
		copyChan <- k
	}
	close(copyChan)

	<-deleteDoneC
	wg2.Wait()

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

func deleteWorkers(chan<- error) <-chan bool {
	doneC := make(chan bool, 0)
	errorC := make(chan error, 0)
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
