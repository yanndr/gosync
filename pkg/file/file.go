package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const bufferSize = 4096

type Copier interface {
	//Copy a sourceFile to the destinationFile, if the parent folder doesn't exist it will be created.
	Copy(sourceFile, destinationFile string, symlink bool) error
}

type BasicCopy struct{}

func (*BasicCopy) Copy(sourceFile, destinationFile string, symlink bool) error {
	if symlink {
		return copySymlink(sourceFile, destinationFile)
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

	buf := make([]byte, bufferSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("cannot read from  buffer for file %s: %w", sourceFile, err)
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return fmt.Errorf("cannot write in buffer for file %s: %w", destinationFile, err)
		}
	}
	return nil
}

func copySymlink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return fmt.Errorf("cannot read symlink %s: %w", source, err)
	}
	err = os.Symlink(link, dest)
	if err != nil {
		return fmt.Errorf("cannot create symlink %s: %w", dest, err)
	}

	return nil
}
