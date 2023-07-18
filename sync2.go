package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
)

type File struct {
	source, destination string
	fileMode            os.FileMode
}

type syncDirectory struct {
	source, destination string
}

type DirectorySync struct {
	source, destination string
	copyC               chan File
	deleteC             chan string
}

func (d *DirectorySync) Sync(maxGoRoutine int) error {
	d.copyC = make(chan File, maxGoRoutine)

	doneC := d.copyListener(maxGoRoutine)

	err := d.sync()
	if err != nil {
		return err
	}
	close(d.copyC)
	<-doneC
	return nil
}

func (d *DirectorySync) copyListener(maxGoroutine int) <-chan interface{} {
	doneC := make(chan interface{})
	go func() {
		wg := sync.WaitGroup{}
		semaphore := make(chan struct{}, maxGoroutine)

		for f := range d.copyC {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(fi File) {
				defer func() {
					wg.Done()
					<-semaphore
				}()
				err := Copy(fi.source, fi.destination)
				if err != nil {
					log.Printf("ERROR:%s", err)
				}
			}(f)

		}

		wg.Wait()
		doneC <- struct{}{}
		close(doneC)
	}()

	return doneC
}

func (d *DirectorySync) sync() error {

	if err := isAValidDirectory(d.source); err != nil {
		return err
	}

	if err := isAValidDirectory(d.destination); err != nil {
		return err
	}

	if d.source == d.destination {
		return fmt.Errorf("error source and destination are the same directory")
	}

	queue := make([]syncDirectory, 1)
	queue[0] = syncDirectory{source: d.source, destination: d.destination}

	for len(queue) > 0 {
		sd := queue[0]
		queue = queue[1:]

		existingEntries, err := loadExistingEntries(sd.destination)
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(sd.source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			isDir, exists := existingEntries[entry.Name()]

			if !exists {
				if entry.IsDir() {
					queue = append(queue, syncDirectory{source: path.Join(sd.source, entry.Name()), destination: path.Join(sd.destination, entry.Name())})
				} else {
					d.copyC <- File{source: path.Join(sd.source, entry.Name()), destination: path.Join(sd.destination, entry.Name())}
				}
				continue
			}

			if isDir != entry.IsDir() {
				err := os.RemoveAll(path.Join(sd.destination, entry.Name()))
				if err != nil {
					return err
				}
			}

			if entry.IsDir() {
				queue = append(queue, syncDirectory{source: path.Join(sd.source, entry.Name()), destination: path.Join(sd.destination, entry.Name())})
			} else {
				d.copyC <- File{source: path.Join(sd.source, entry.Name()), destination: path.Join(sd.destination, entry.Name())}
			}

			delete(existingEntries, entry.Name())

		}
		for name := range existingEntries {
			err := os.RemoveAll(path.Join(sd.destination, name))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func loadExistingEntries(folderPath string) (map[string]bool, error) {
	existingEntries := make(map[string]bool)
	destEntries, err := os.ReadDir(folderPath)
	if err != nil {
		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) {
			return nil, fmt.Errorf("cannot read directory %s: %w", folderPath, err)
		}
	} else {
		for _, entry := range destEntries {
			existingEntries[entry.Name()] = entry.IsDir()
		}
	}
	return existingEntries, nil
}

func Copy(sourceFile, destinationFile string) error {
	source, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("cannot open source file %s: %w", sourceFile, err)
	}
	defer source.Close()

	destinationDir := filepath.Dir(destinationFile)
	_, err = os.Stat(destinationDir)
	if err != nil {
		if os.IsNotExist(err) {
			srcDir := filepath.Dir(sourceFile)
			dirStat, err := os.Stat(srcDir)
			err = os.MkdirAll(destinationDir, dirStat.Mode())
			if err != nil {
				return fmt.Errorf("error creating directory %s: %w", destinationDir, err)
			}
		} else {
			return fmt.Errorf("error getting stat for directory %s: %w", destinationDir, err)
		}
	}

	destination, err := os.Create(destinationFile)
	if err != nil {
		return fmt.Errorf("cannot create destination file %s: %w", destinationFile, err)
	}
	defer destination.Close()

	buf := make([]byte, 4096)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

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
