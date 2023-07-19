package tests

import (
	"fmt"
	"gosync/pkg/directory"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	sourceA = "./source_folder_a"
	sourceB = "./source_folder_b"
	sourceC = "./source_folder_c"
	dest    = "./dest_temp"
)

func Test_syncFolderA(t *testing.T) {
	//setup
	defer os.RemoveAll(dest)
	ds := directory.NewSynchronizer(sourceA, dest)

	//act
	err := ds.Sync()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	//verify
	destEntries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(destEntries) != 3 {
		t.Fatalf("expected 3 file got %v file(s)", len(destEntries))
	}
	err = folderMustContains(dest, []string{dest, path.Join(dest, "file_a"), path.Join(dest, "file_b"), path.Join(dest, "file_c")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_syncFolderAThenB(t *testing.T) {
	//setup
	defer os.RemoveAll(dest)
	ds := directory.NewSynchronizer(sourceA, dest)
	ds2 := directory.NewSynchronizer(sourceB, dest)

	//act
	err := ds.Sync()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = ds2.Sync()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	//verify
	destEntries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(destEntries) != 3 {
		t.Fatalf("expected 3 file got %v file(s)", len(destEntries))
	}
	err = folderMustContains(dest, []string{dest, path.Join(dest, "file_a"), path.Join(dest, "file_c"), path.Join(dest, "file_d")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_syncFolderBThenC(t *testing.T) {
	//setup
	defer os.RemoveAll(dest)
	ds := directory.NewSynchronizer(sourceB, dest)
	ds2 := directory.NewSynchronizer(sourceC, dest)

	//act
	err := ds.Sync()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = ds2.Sync()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	//verify
	destEntries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(destEntries) != 4 {
		t.Fatalf("expected 3 entries got %v entries", len(destEntries))
	}
	err = folderMustContains(dest, []string{dest, path.Join(dest, "dir_a"), path.Join(dest, "dir_a", "file_a_a"), path.Join(dest, "file_a"), path.Join(dest, "file_d"), path.Join(dest, "file_e")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func folderMustContains(folderPath string, expected []string) error {
	files := make([]string, 0)
	err := filepath.Walk(folderPath, func(path string, info fs.FileInfo, err error) error {
		fmt.Println(path)
		files = append(files, path)
		return err
	})

	if err != nil {
		return err
	}

	if !reflect.DeepEqual(files, expected) {
		return fmt.Errorf("expected files %v, got %v", expected, files)
	}

	return nil
}
