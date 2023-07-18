package main

import (
	"errors"
	"fmt"
	"gosync/file"
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

type DirectorySync struct {
	source, destination string
	copyC               chan File
	semaphore           chan struct{}
}

func (d *DirectorySync) Sync() error {
	d.copyC = make(chan File, 20)
	d.semaphore = make(chan struct{}, 20)
	doneC := make(chan interface{})
	go d.copyListener(doneC)

	err := d.sync()
	if err != nil {
		return err
	}
	close(d.copyC)
	<-doneC
	return nil
}

func (d *DirectorySync) copyListener(doneC chan<- interface{}) {
	wg := sync.WaitGroup{}
	for f := range d.copyC {
		wg.Add(1)
		d.semaphore <- struct{}{}
		go func(fi File) {
			defer wg.Done()
			err := file.Copy(fi.source, fi.destination)
			if err != nil {
				log.Printf("ERROR:%s", err)
			}
		}(f)
		<-d.semaphore
	}
	go func() {
		wg.Wait()
		doneC <- struct{}{}
	}()
}

type syncDirectories struct {
	source, destination string
}

func (d *DirectorySync) sync() error {
	queue := make([]syncDirectories, 1)
	queue[0] = syncDirectories{source: d.source, destination: d.destination}

	for len(queue) > 0 {
		sd := queue[0]
		queue = queue[1:]
		existingEntries := make(map[string]bool)
		destEntries, err := os.ReadDir(sd.destination)
		if err != nil {
			var pathErr *fs.PathError
			if !errors.As(err, &pathErr) {
				return fmt.Errorf("cannot read directory %s: %w", d.destination, err)
			}
		} else {
			for _, entry := range destEntries {
				existingEntries[entry.Name()] = entry.IsDir()
			}
		}

		entries, err := os.ReadDir(sd.source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			isDirectory, ok := existingEntries[entry.Name()]
			if !ok {
				if entry.IsDir() {
					queue = append(queue, syncDirectories{source: path.Join(sd.source, entry.Name()), destination: path.Join(sd.destination, entry.Name())})
				} else {
					d.copyC <- File{source: path.Join(sd.source, entry.Name()), destination: path.Join(sd.destination, entry.Name())}
				}
				delete(existingEntries, entry.Name())
			} else {
				if isDirectory {
					queue = append(queue, syncDirectories{source: path.Join(sd.source, entry.Name()), destination: path.Join(sd.destination, entry.Name())})
				}
				delete(existingEntries, entry.Name())

			}
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

// copyDirectory copy the whole source directory to the destination directory.
func (d *DirectorySync) copyDirectory(source, destination string) error {

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destination, relativePath)
		if info.IsDir() {
			err := os.MkdirAll(destPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			d.copyC <- File{source: path, destination: destPath, fileMode: info.Mode()}
		}
		return nil
	})
	return err
}
