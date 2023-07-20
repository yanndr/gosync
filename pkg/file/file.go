package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Copier interface {
	//Copy a sourceFile to the destinationFile, if the parent folder doesn't exist it will be created
	Copy(sourceFile, destinationFile string, symlink bool) error
}

type BasicCopy struct{}

func (*BasicCopy) Copy(sourceFile, destinationFile string, symlink bool) error {
	if symlink {
		return copySymLink(sourceFile, destinationFile)
	}

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

func copySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
