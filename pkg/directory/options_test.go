package directory

import (
	"testing"
)

func Test_copyBufferSize(t *testing.T) {
	options := defaultOptions

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
			CopyBufferSize(tt.size).apply(&options)
			if options.copyBufferSize != tt.want {
				t.Errorf("CopyBufferSize() = %v, want %v", options.copyBufferSize, tt.want)
			}
		})
	}
}

func Test_maxGoroutine(t *testing.T) {
	options := defaultOptions

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
			MaxGoroutine(tt.max).apply(&options)
			if options.maxGoroutine != tt.want {
				t.Errorf("maxGoroutine() = %v, want %v", options.maxGoroutine, tt.want)
			}
		})
	}
}

func Test_fileCopier(t *testing.T) {
	options := defaultOptions
	fk := &fakeCopier{}
	FileCopier(fk).apply(&options)
	if options.fileCopier != fk {
		t.Errorf("fileCopier() = %v, want %v", options.fileCopier, fk)
	}

}
