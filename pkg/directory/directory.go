package directory

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type EntryType int

const (
	folder = EntryType(iota)
	file
	symlink
)

type InputError struct {
	msg string
}

func (e *InputError) Error() string {
	return fmt.Sprintf("input error - %s", e.msg)
}

// IsValid returns an error if the path doesn't exist, or it is not a directory
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
	// ListEntries list all the entries in the folderPath and returns a map[string]EntryType of the entries
	listEntries(folderPath string) (map[string]EntryType, error)
}

type basicDirEntryLister struct {
}

func (basicDirEntryLister) listEntries(folderPath string) (map[string]EntryType, error) {
	return ListEntries(folderPath)
}

// ListEntries list all the entries in the folderPath and returns a map[string]EntryType of the entries
func ListEntries(folderPath string) (map[string]EntryType, error) {
	existingEntries := make(map[string]EntryType)
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

func getEntryType(fileMode os.FileMode) EntryType {
	switch fileMode {
	case os.ModeDir:
		return folder
	case os.ModeSymlink:
		return symlink
	default:
		return file
	}
}
