package gosync

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

type entryType bool

const (
	folder = entryType(true)
	file   = entryType(false)
)

type File struct {
	source, destination string
	fileMode            os.FileMode
}

const (
	defaultMaxGoroutine = 20
)

type DirectorySynchronizer struct {
	Source, Destination string
	copyC               chan File
	deleteC             chan string
	maxGoroutine        int
	copyBuffer          int
}

func (d *DirectorySynchronizer) Sync() error {
	if err := isAValidDirectory(d.Source); err != nil {
		return err
	}

	if d.Source == d.Destination {
		return fmt.Errorf("error Source and Destination are the same directory")
	}

	if d.maxGoroutine == 0 {
		d.maxGoroutine = defaultMaxGoroutine
	}
	if d.copyBuffer == 0 {
		d.copyBuffer = defaultMaxGoroutine * 2
	}
	if d.copyC == nil {
		d.copyC = make(chan File, d.copyBuffer)
	}

	doneC, errorC := d.copyListener(d.maxGoroutine)

	errorOccurred := false
	go func() {
		for err := range errorC {
			errorOccurred = true
			log.Println(err)
		}
	}()

	err := d.synchronizeFolder()
	if err != nil {
		return fmt.Errorf("cannot perform the synchronization: %w", err)
	}
	close(d.copyC)
	<-doneC

	if errorOccurred {
		return fmt.Errorf("process ended with copy errors")
	}
	return nil
}

func (d *DirectorySynchronizer) copyListener(maxGoroutine int) (<-chan interface{}, <-chan error) {
	doneC := make(chan interface{})
	errorC := make(chan error)
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
				err := copyFile(fi.source, fi.destination)
				if err != nil {
					errorC <- err
				}
			}(f)

		}

		wg.Wait()
		doneC <- struct{}{}
		close(doneC)
	}()

	return doneC, errorC
}

func (d *DirectorySynchronizer) synchronizeFolder() error {

	type syncFolders struct {
		source, destination string
	}

	folderQueue := make([]syncFolders, 1)
	folderQueue[0] = syncFolders{source: d.Source, destination: d.Destination}

	for len(folderQueue) > 0 {
		folders := folderQueue[0]
		folderQueue = folderQueue[1:]

		existingEntries, err := listExistingEntries(folders.destination)
		if err != nil {
			return fmt.Errorf("cannot load entries from %s: %w", folders.destination, err)
		}

		entries, err := os.ReadDir(folders.source)
		if err != nil {
			return fmt.Errorf("cannot read directory %s: %w", folders.source, err)
		}
		for _, entry := range entries {
			isDir, exists := existingEntries[entry.Name()]

			if !exists {
				if entry.IsDir() {
					folderQueue = append(folderQueue, syncFolders{source: path.Join(folders.source, entry.Name()), destination: path.Join(folders.destination, entry.Name())})
				} else {
					d.copyC <- File{source: path.Join(folders.source, entry.Name()), destination: path.Join(folders.destination, entry.Name())}
				}
				continue
			}

			if isDir != entryType(entry.IsDir()) {
				err = os.RemoveAll(path.Join(folders.destination, entry.Name()))
				if err != nil {
					return fmt.Errorf("cannot delete entry %s: %w", entry.Name(), err)
				}
			}

			if entry.IsDir() {
				folderQueue = append(folderQueue, syncFolders{source: path.Join(folders.source, entry.Name()), destination: path.Join(folders.destination, entry.Name())})
			} else {
				d.copyC <- File{source: path.Join(folders.source, entry.Name()), destination: path.Join(folders.destination, entry.Name())}
			}

			delete(existingEntries, entry.Name())

		}
		for name := range existingEntries {
			err = os.RemoveAll(path.Join(folders.destination, name))
			if err != nil {
				return fmt.Errorf("cannot delete entry %s: %w", name, err)
			}
		}
	}

	return nil
}

// listExistingEntries list all the entries in the folderPath and returns a map[string]bool of the entries
func listExistingEntries(folderPath string) (map[string]entryType, error) {
	existingEntries := make(map[string]entryType)
	destEntries, err := os.ReadDir(folderPath)
	if err != nil {
		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) {
			return nil, fmt.Errorf("cannot read directory %s: %w", folderPath, err)
		}
	} else {
		for _, entry := range destEntries {
			existingEntries[entry.Name()] = entryType(entry.IsDir())
		}
	}
	return existingEntries, nil
}

type fileCopier interface {
	copyFile(sourceFile, destinationFile string) error
}

// copyFile copy a sourceFile to the destinationFile, if the parent folder doesn't exist it will be created
func copyFile(sourceFile, destinationFile string) error {
	source, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("cannot open Source file %s: %w", sourceFile, err)
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
		return fmt.Errorf("cannot create Destination file %s: %w", destinationFile, err)
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

// IsAValidDirectory returns an error if the path doesn't exist, or it is not a directory
func isAValidDirectory(path string) error {
	sourceInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s is not a valid directory: %w", path, err)
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}
