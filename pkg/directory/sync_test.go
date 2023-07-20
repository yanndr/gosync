package directory

import (
	"sync"
	"testing"
)

type fakeCopier struct {
	fileCopied int
	mu         sync.Mutex
}

func (c *fakeCopier) Copy(source, destination string, symlink bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fileCopied++
	return nil
}

func Test_synchronizer_Sync(t *testing.T) {

	type fields struct {
		Source      string
		Destination string
	}
	tests := []struct {
		name           string
		fields         fields
		wantFileCopied int
		wantErr        bool
	}{
		{"same directory", fields{"a", "a"}, 0, true},
		{"folder_a", fields{"../../tests/source_folder_a", "a"}, 3, false},
		{"folder_a", fields{"../../tests/source_folder_c", "a"}, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &fakeCopier{mu: sync.Mutex{}}
			s := NewSynchronizer(tt.fields.Source, tt.fields.Destination, fileCopier(fc))
			if err := s.Sync(); (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantFileCopied != fc.fileCopied {
				t.Errorf("Sync() file copied = %v, want %v", fc.fileCopied, tt.wantFileCopied)
			}
		})
	}
}

type fakeEntryLister struct {
	result map[string]entryType
}

func (el *fakeEntryLister) listEntries(folder string) (map[string]entryType, error) {
	return el.result, nil
}

func Test_synchronizer_Sync_withExistingFiles(t *testing.T) {

	el := &fakeEntryLister{}

	type fields struct {
		Source      string
		Destination string
		files       map[string]entryType
	}
	tests := []struct {
		name           string
		fields         fields
		wantFileCopied int
		wantErr        bool
	}{
		{"one file already on dest", fields{"../../tests/source_folder_a", "a", map[string]entryType{"file_a": file}}, 2, false},
		{"all files already on dest", fields{"../../tests/source_folder_a", "a", map[string]entryType{"file_a": file, "file_b": file, "file_c": file}}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el.result = tt.fields.files
			fc := &fakeCopier{mu: sync.Mutex{}}
			s2 := NewSynchronizer(tt.fields.Source, tt.fields.Destination, fileCopier(fc), entryLister(el))
			if err := s2.Sync(); (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantFileCopied != fc.fileCopied {
				t.Errorf("Sync() file copied = %v, want %v", fc.fileCopied, tt.wantFileCopied)
			}
		})
	}
}
