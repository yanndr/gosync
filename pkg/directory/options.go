package directory

import syncFile "gosync/pkg/file"

const (
	defaultMaxGoroutine   = 20
	defaultCopyBufferSize = 20
)

// SynchronizerOption sets options such as  MaxGoroutine and CopyBufferSize
type SynchronizerOption interface {
	apply(option *synchronizer)
}

var defaultSynchronizer = synchronizer{
	maxGoroutine:   defaultMaxGoroutine,
	copyBufferSize: defaultCopyBufferSize,
	fileCopier:     &syncFile.BasicCopy{},
	entryLister:    &basicDirEntryLister{},
}

type funcSynchronizerOption struct {
	f func(*synchronizer)
}

func (fo *funcSynchronizerOption) apply(s *synchronizer) {
	fo.f(s)
}

func newFuncSynchronizerOption(f func(*synchronizer)) *funcSynchronizerOption {
	return &funcSynchronizerOption{
		f: f,
	}
}

// MaxGoroutine lets you set the maximum number of goroutines that are allowed for copying files.
func MaxGoroutine(m int) SynchronizerOption {
	return newFuncSynchronizerOption(func(s *synchronizer) {
		if m > 0 {
			s.maxGoroutine = m
		}

	})
}

// CopyBufferSize lets you set the copy channel buffer size.
func CopyBufferSize(size int) SynchronizerOption {
	return newFuncSynchronizerOption(func(s *synchronizer) {
		if size > 0 {
			s.copyBufferSize = size
		}
	})
}

// fileCopier lets you set up the syncFile.Copier for testing purpose.
func fileCopier(fc syncFile.Copier) SynchronizerOption {
	return newFuncSynchronizerOption(func(s *synchronizer) {
		s.fileCopier = fc
	})
}

// entryLister lets you set up the dirEntryLister for testing purpose.
func entryLister(el dirEntryLister) SynchronizerOption {
	return newFuncSynchronizerOption(func(o *synchronizer) {
		o.entryLister = el
	})
}
