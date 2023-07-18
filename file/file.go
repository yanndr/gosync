package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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

	/*

		bufferSize := 4096 // Adjust the buffer size based on your system and requirements
		buffer := make([]byte, bufferSize)

		// Get the file size
		sourceInfo, err := source.Stat()
		if err != nil {
			return fmt.Errorf("cannot get source file %s info: %w", sourceFile, err)
		}
		fileSize := sourceInfo.Size()

		// Calculate the number of chunks and remaining bytes
		numChunks := fileSize / int64(bufferSize)
		remainingBytes := fileSize % int64(bufferSize)

		// Start goroutines to copy the chunks in parallel
		done := make(chan bool)
		errors := make(chan error)

		for i := int64(0); i < numChunks; i++ {
			go func(offset int64) {
				_, err := source.ReadAt(buffer, offset)
				if err != nil && err != io.EOF {
					errors <- err
					return
				}

				_, err = destination.Write(buffer)
				if err != nil {
					errors <- err
					return
				}

				done <- true
			}(i * int64(bufferSize))
		}

		// Copy the remaining bytes in the main goroutine
		_, err = io.CopyN(destination, source, remainingBytes)
		if err != nil && err != io.EOF {
			return err
		}

		// Wait for all goroutines to finish
		for i := int64(0); i < numChunks; i++ {
			<-done
		}

		// Check for any errors reported by the goroutines
		select {
		case err := <-errors:
			return err
		default:
		}
	*/
	return nil
}
