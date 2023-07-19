package directory

import (
	"os"
	"reflect"
	"testing"
)

func TestIsValid(t *testing.T) {
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
			if err := IsValid(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("isAValidDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListExistingEntries(t *testing.T) {
	tests := []struct {
		name       string
		folderPath string
		want       map[string]EntryType
		wantErr    bool
	}{
		{"Empty path", "", map[string]EntryType{}, false},
		{"folder a path", "../../tests/source_folder_a", map[string]EntryType{"file_a": file, "file_b": file, "file_c": file}, false},
		{"folder c path", "../../tests/source_folder_c", map[string]EntryType{"dir_a": folder, "file_a": file, "file_d": file, "file_e": file}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListEntries(tt.folderPath)
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
