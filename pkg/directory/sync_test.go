package directory

import (
	"sync"
	"testing"
)

type fakeCopier struct {
	fileCopied int
	mu         sync.Mutex
}

func (c *fakeCopier) Copy(source, destination string) error {
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
			s := NewSynchronizer(tt.fields.Source, tt.fields.Destination, FileCopier(fc))
			if err := s.Sync(); (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantFileCopied != fc.fileCopied {
				t.Errorf("Sync() file copied = %v, want %v", fc.fileCopied, tt.wantFileCopied)
			}
		})
	}
}
