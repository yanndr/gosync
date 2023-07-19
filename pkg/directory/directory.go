package directory

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type EntryType bool

const (
	folder = EntryType(true)
	file   = EntryType(false)
)

// IsValid returns an error if the path doesn't exist, or it is not a directory
func IsValid(path string) error {
	sourceInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s is not a valid directory: %w", path, err)
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}

// ListExistingEntries list all the entries in the folderPath and returns a map[string]bool of the entries
func ListExistingEntries(folderPath string) (map[string]EntryType, error) {
	existingEntries := make(map[string]EntryType)
	destEntries, err := os.ReadDir(folderPath)
	if err != nil {
		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) {
			return nil, fmt.Errorf("cannot read directory %s: %w", folderPath, err)
		}
	} else {
		for _, entry := range destEntries {
			existingEntries[entry.Name()] = EntryType(entry.IsDir())
		}
	}
	return existingEntries, nil
}
