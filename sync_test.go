package gosync

import (
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_isAValidDirectory(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		t.Fatal("cannot create temp dir for test")
	}
	defer os.RemoveAll(dir)

	file, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		t.Fatal("cannot create temp dir for test")
	}
	defer os.Remove(file.Name())
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"empty path", "", true},
		{"invalid path", "?e", true},
		{"non existing path", "./idonotexist", true},
		{"existing dir", dir, false},
		{"existing file", file.Name(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := isAValidDirectory(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("isAValidDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_copyFile(t *testing.T) {
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
		{"empty source", args{"./tests/source_folder_a/file_a", path.Join(dir, "file_a")}, false},
		{"non existing file", args{"./tests/source_folder_a/file_e", path.Join(dir, "file_a")}, true},
		{"existing file", args{"./tests/source_folder_a/file_a", path.Join(dir, "file_a")}, false},
		{"existing file in a sub folder", args{"./tests/source_folder_c/dir_a/file_a_a", path.Join(dir, "/dir_a/file_a_a")}, false},
		{"existing file with non existing destination folder", args{"./tests/source_folder_a/file_a", newTempDir}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := copyFile(tt.args.sourceFile, tt.args.destinationFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("copyFile() error = %v, wantErr %v", err, tt.wantErr)
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

func Test_listExistingEntries(t *testing.T) {
	tests := []struct {
		name       string
		folderPath string
		want       map[string]bool
		wantErr    bool
	}{
		{
			name:       "",
			folderPath: "",
			want:       nil,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listExistingEntries(tt.folderPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("listExistingEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listExistingEntries() got = %v, want %v", got, tt.want)
			}
		})
	}
}
