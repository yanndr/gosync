package directory

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"
)

type File struct {
	source, destination string
	fileMode            os.FileMode
}

type Synchronizer interface {
	Sync() error
}

type synchronizer struct {
	Source, Destination string
	copyC               chan File
	deleteC             chan string
	options             synchronizerOption
}

func NewSynchronizer(source, destination string, opts ...SynchronizerOption) Synchronizer {
	options := defaultOptions
	if opts != nil {
		for _, opt := range opts {
			opt.apply(&options)
		}
	}
	s := &synchronizer{
		Source:      source,
		Destination: destination,
		options:     options,
		copyC:       make(chan File, options.copyBufferSize),
	}

	return s
}

func (s *synchronizer) Sync() error {
	if err := IsValid(s.Source); err != nil {
		return err
	}

	if s.Source == s.Destination {
		return fmt.Errorf("error Source and Destination are the same directory")
	}

	doneC, errorC := s.copyListener(s.options.maxGoroutine)

	errorOccurred := false
	go func() {
		for err := range errorC {
			errorOccurred = true
			log.Println(err)
		}
	}()

	err := s.synchronizeFolder()
	if err != nil {
		return fmt.Errorf("cannot perform the synchronization: %w", err)
	}
	close(s.copyC)
	<-doneC

	if errorOccurred {
		return fmt.Errorf("process ended with copy errors")
	}
	return nil
}

func (s *synchronizer) copyListener(maxGoroutine int) (<-chan interface{}, <-chan error) {
	doneC := make(chan interface{})
	errorC := make(chan error)
	go func() {
		wg := sync.WaitGroup{}
		semaphore := make(chan struct{}, maxGoroutine)

		for f := range s.copyC {
			wg.Add(1)
			semaphore <- struct{}{}
			go func(fi File) {
				defer func() {
					wg.Done()
					<-semaphore
				}()
				err := s.options.fileCopier.Copy(fi.source, fi.destination)
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

func (s *synchronizer) synchronizeFolder() error {

	type syncFolders struct {
		source, destination string
	}

	folderQueue := make([]syncFolders, 1)
	folderQueue[0] = syncFolders{source: s.Source, destination: s.Destination}

	for len(folderQueue) > 0 {
		folders := folderQueue[0]
		folderQueue = folderQueue[1:]

		existingEntries, err := ListExistingEntries(folders.destination)
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
					s.copyC <- File{source: path.Join(folders.source, entry.Name()), destination: path.Join(folders.destination, entry.Name())}
				}
				continue
			}

			if isDir != EntryType(entry.IsDir()) {
				err = os.RemoveAll(path.Join(folders.destination, entry.Name()))
				if err != nil {
					return fmt.Errorf("cannot delete entry %s: %w", entry.Name(), err)
				}
			}

			if entry.IsDir() {
				folderQueue = append(folderQueue, syncFolders{source: path.Join(folders.source, entry.Name()), destination: path.Join(folders.destination, entry.Name())})
			} else {
				s.copyC <- File{source: path.Join(folders.source, entry.Name()), destination: path.Join(folders.destination, entry.Name())}
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
