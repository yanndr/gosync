package directory

import (
	"fmt"
	syncFile "gosync/pkg/file"
	"os"
	"path"
	"strings"
	"sync"
)

type CopyError struct {
	errors []string
}

func (e *CopyError) Error() string {
	return strings.Join(e.errors, "\n")
}

// fileSync is a pair of path of file or symlink to synchronize.
type fileSync struct {
	source, destination string
	fileType            entryType
}

// Synchronizer is a directory synchronizer between a source and a destination folder.
type Synchronizer interface {
	//Sync launch the syncing operation between the two folders
	Sync() error
}

type synchronizer struct {
	Source, Destination string
	copyC               chan fileSync
	maxGoroutine        int
	copyBufferSize      int
	fileCopier          syncFile.Copier
	entryLister         dirEntryLister
}

// NewSynchronizer initialize a directory synchronizer.
func NewSynchronizer(source, destination string, opts ...SynchronizerOption) Synchronizer {
	s := defaultSynchronizer
	if opts != nil {
		for _, opt := range opts {
			opt.apply(&s)
		}
	}
	s.Source = source
	s.Destination = destination
	s.copyC = make(chan fileSync, s.copyBufferSize)

	return &s
}

func (s *synchronizer) Sync() error {
	if err := IsValid(s.Source); err != nil {
		return err
	}

	if s.Source == s.Destination {
		return &InputError{msg: "error: Source and Destination are the same directory"}
	}

	doneC, errorC := s.copyListener(s.maxGoroutine)

	errs := make([]string, 0)
	go func() {
		for err := range errorC {
			errs = append(errs, err.Error())
		}
	}()

	err := s.synchronizeFolder()
	if err != nil {
		return fmt.Errorf("cannot perform the synchronization: %w", err)
	}
	close(s.copyC)
	<-doneC

	if len(errs) > 0 {
		return &CopyError{errors: errs}
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
			go func(fi fileSync) {
				defer func() {
					wg.Done()
					<-semaphore
				}()
				err := s.fileCopier.Copy(fi.source, fi.destination, fi.fileType == symlink)
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

		existingEntries, err := s.entryLister.listEntries(folders.destination)
		if err != nil {
			return fmt.Errorf("cannot load entries from %s: %w", folders.destination, err)
		}

		entries, err := os.ReadDir(folders.source)
		if err != nil {
			return fmt.Errorf("cannot read directory %s: %w", folders.source, err)
		}
		for _, entry := range entries {
			destEntryType, exists := existingEntries[entry.Name()]
			sourceEntryType := getEntryType(entry.Type())

			source := path.Join(folders.source, entry.Name())
			destination := path.Join(folders.destination, entry.Name())

			if !exists {
				if sourceEntryType == file || sourceEntryType == symlink {
					s.copyC <- fileSync{source: source, destination: destination, fileType: sourceEntryType}
				}
			} else {
				if destEntryType != sourceEntryType {
					err = os.RemoveAll(destination)
					if err != nil {
						return fmt.Errorf("cannot delete entry %s: %w", entry.Name(), err)
					}

					if sourceEntryType == file || sourceEntryType == symlink {
						s.copyC <- fileSync{source: source, destination: destination, fileType: sourceEntryType}
					}
				}
				delete(existingEntries, entry.Name())
			}

			if sourceEntryType == folder {
				folderQueue = append(folderQueue, syncFolders{source: source, destination: destination})
			}
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
