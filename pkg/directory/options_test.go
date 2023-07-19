package directory

import (
	"testing"
)

func Test_copyBufferSize(t *testing.T) {
	s := defaultSynchronizer

	tests := []struct {
		name string
		size int
		want int
	}{
		{"zero", 0, defaultCopyBufferSize},
		{"1", 1, 1},
		{"10", 10, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CopyBufferSize(tt.size).apply(&s)
			if s.copyBufferSize != tt.want {
				t.Errorf("CopyBufferSize() = %v, want %v", s.copyBufferSize, tt.want)
			}
		})
	}
}

func Test_maxGoroutine(t *testing.T) {
	s := defaultSynchronizer

	tests := []struct {
		name string
		max  int
		want int
	}{
		{"zero", 0, defaultMaxGoroutine},
		{"1", 1, 1},
		{"10", 10, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MaxGoroutine(tt.max).apply(&s)
			if s.maxGoroutine != tt.want {
				t.Errorf("maxGoroutine() = %v, want %v", s.maxGoroutine, tt.want)
			}
		})
	}
}

func Test_fileCopier(t *testing.T) {
	s := defaultSynchronizer
	fk := &fakeCopier{}
	fileCopier(fk).apply(&s)
	if s.fileCopier != fk {
		t.Errorf("fileCopier() = %v, want %v", s.fileCopier, fk)
	}

}
