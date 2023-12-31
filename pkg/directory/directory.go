package directory

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type entryType int

const (
	folder = entryType(iota)
	file
	symlink
)

type InputError struct {
	msg string
}

func (e *InputError) Error() string {
	return fmt.Sprintf("input error - %s", e.msg)
}

// IsValid returns an error if the path doesn't exist, or if it is not a directory
func IsValid(path string) error {
	sourceInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s is not a valid directory: %w", path, err)
	}
	if !sourceInfo.IsDir() {
		return &InputError{msg: fmt.Sprintf("%s is not a directory:", path)}
	}

	return nil
}

type dirEntryLister interface {
	// listEntries lists all the entries in the folderPath and returns a map[string]entryType of the entries.
	listEntries(folderPath string) (map[string]entryType, error)
}

type basicDirEntryLister struct {
}

func (basicDirEntryLister) listEntries(folderPath string) (map[string]entryType, error) {
	return ListEntries(folderPath)
}

// ListEntries lists all the entries in the folderPath and returns a map[string]entryType of the entries.
func ListEntries(folderPath string) (map[string]entryType, error) {
	existingEntries := make(map[string]entryType)
	destEntries, err := os.ReadDir(folderPath)
	if err != nil {
		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) {
			return nil, fmt.Errorf("cannot read directory %s: %w", folderPath, err)
		}
	} else {
		for _, entry := range destEntries {
			existingEntries[entry.Name()] = getEntryType(entry.Type())
		}
	}
	return existingEntries, nil
}

func getEntryType(fileMode os.FileMode) entryType {
	switch fileMode {
	case os.ModeDir:
		return folder
	case os.ModeSymlink:
		return symlink
	default:
		return file
	}
}
