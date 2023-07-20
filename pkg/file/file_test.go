package file

import (
	"os"
	"path"
	"testing"
)

func TestBasicCopy_Copy(t *testing.T) {

	dir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		t.Fatal("cannot create temp dir for test")
	}
	defer os.RemoveAll(dir)
	newTempDir := path.Join(os.TempDir(), "tempgosync")
	defer os.RemoveAll(newTempDir)

	type args struct {
		sourceFile      string
		destinationFile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty path", args{"", ""}, true},
		{"empty source", args{"../../tests/source_folder_a/file_a", path.Join(dir, "file_a")}, false},
		{"non existing file", args{"../../tests/source_folder_a/file_e", path.Join(dir, "file_a")}, true},
		{"existing file", args{"../../tests/source_folder_a/file_a", path.Join(dir, "file_a")}, false},
		{"existing file in a sub folder", args{"../../tests/source_folder_c/dir_a/file_a_a", path.Join(dir, "/dir_a/file_a_a")}, false},
		{"existing file with non existing destination folder", args{"../../tests/source_folder_a/file_a", newTempDir}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := BasicCopy{}
			err := ba.Copy(tt.args.sourceFile, tt.args.destinationFile, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				_, err := os.Stat(tt.args.destinationFile)
				if err != nil {
					t.Errorf("copyFile()  file %s should exists, %v", tt.args.destinationFile, err)
				}
			}
		})
	}
}
