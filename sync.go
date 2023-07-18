package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
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

func synchronize(source, destination string) error {

	var filesToSync = make(map[string]os.FileMode)
	var directoryToDelete = make(map[string]bool)

	if err := isAValidDirectory(source); err != nil {
		return err
	}

	if err := isAValidDirectory(destination); err != nil {
		return err
	}

	if source == destination {
		return fmt.Errorf("error source and dest are the same directory")
	}

	errorC := make(chan error)

	errors := false
	go func() {
		for err := range errorC {
			errors = true
			log.Println(err)
		}
	}()

	err := walk(source, "", func(filePath string, fileMode os.FileMode) {
		filesToSync[filePath] = fileMode
	}, func(path string) {
		directoryToDelete[path] = true
	})

	if err != nil {
		return err
	}

	deleteDoneC := workers(numberOfWorker, deleteChan, errorC, func(path string) error {
		err := os.Remove(path)
		if err != nil {
			return err
		}
		return nil
	})

	err = walk(destination, "", func(filePath string, _ os.FileMode) {
		if _, ok := filesToSync[filePath]; ok {
			delete(filesToSync, filePath)
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

	copyDoneC := workers(numberOfWorker, copyChan, errorC, copyFileFn(source, destination, filesToSync))

	for k := range filesToSync {
		copyChan <- k
	}
	close(copyChan)

	<-deleteDoneC
	<-copyDoneC

	if errors {
		return fmt.Errorf("operation complete with errors")
	}
	return nil
}

func walk(baseDir, subDir string, fileFn func(path string, fileMode os.FileMode), directoryFn func(path string)) error {
	workingDir := path.Join(baseDir, subDir)
	entries, err := os.ReadDir(workingDir)
	if err != nil {
		return fmt.Errorf("error listing files in directory %s: %w", workingDir, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			directoryFn(path.Join(subDir, e.Name()))

			err = walk(baseDir, path.Join(subDir, e.Name()), fileFn, directoryFn)
			if err != nil {
				return fmt.Errorf("error scannig directory, %s: %w", workingDir, err)
			}
		} else {
			file := path.Join(workingDir, e.Name())
			st, err := os.Stat(file)
			if err != nil {
				return fmt.Errorf("error getting stat for file %s: %w", file, err)
			}
			fileFn(path.Join(subDir, e.Name()), st.Mode())
		}
	}

	return nil
}

func workers(numberOfWorker int, inputChan <-chan string, errorC chan<- error, fn func(path string) error) <-chan bool {
	doneC := make(chan bool, 0)
	wg := sync.WaitGroup{}
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

func copyFileFn(sourceDir, destinationDir string, fileToSync map[string]os.FileMode) func(filePath string) error {
	return func(filePath string) error {
		return copyFile(sourceDir, destinationDir, filePath, fileToSync[filePath])
	}
}

func copyFile(sourceDir, destinationDir, filePath string, fileMode os.FileMode) error {
	sourceFile := path.Join(sourceDir, filePath)
	src, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("error reading source file %s: %w", sourceFile, err)
	}
	defer src.Close()

	destinationFile := path.Join(destinationDir, filePath)
	dir := filepath.Dir(destinationFile)

	_, err = os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			srcDir := filepath.Dir(sourceFile)
			dirStat, err := os.Stat(srcDir)
			err = os.MkdirAll(dir, dirStat.Mode())
			if err != nil {
				return fmt.Errorf("error creating directory %s: %w", dir, err)
			}
		} else {
			return fmt.Errorf("error getting stat for directory %s: %w", dir, err)
		}
	}

	destination, err := os.OpenFile(destinationFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", destinationFile, err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, src)
	if err != nil {
		return fmt.Errorf("error copying file %s to %s: %w", sourceFile, destinationFile, err)
	}

	return nil
}

//type FolderSynchronizer struct {
//	sourceFolder, destinationFolder string
//	copyChanel                      chan file
//}

func syncFolders(sourceFolder, destinationFolder string) error {

	err := deleteNonExistingFolder(sourceFolder, destinationFolder)
	if err != nil {
		return err
	}

	return copyingNewfiles(sourceFolder, destinationFolder)
}

func copyingNewfiles(sourceFolder string, destinationFolder string) error {
	err := filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(sourceFolder, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destinationFolder, relativePath)
		if info.IsDir() {
			err := os.MkdirAll(destPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			destInfo, err := os.Stat(destPath)
			if os.IsNotExist(err) || destInfo.IsDir() {
				err = copyF(path, destPath)
				if err != nil {
					return err
				}
				fmt.Println("Copied:", path)
			}
		}
		return nil
	})

	return err
}

func deleteNonExistingFolder(sourceFolder string, destinationFolder string) error {
	err := filepath.Walk(destinationFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(destinationFolder, path)
		if err != nil {
			return err
		}
		sourcePath := filepath.Join(sourceFolder, relativePath)
		if !fileExists(sourcePath) {
			if info.IsDir() {
				err := os.RemoveAll(path)
				if err != nil {
					return err
				}
				fmt.Println("Deleted folder:", path)
			} else {
				err := os.Remove(path)
				if err != nil {
					return err
				}
				fmt.Println("Deleted file:", path)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func copyF(sourcePath, destPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return err
	}

	return dest.Sync()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
