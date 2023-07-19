package directory

import syncFile "gosync/pkg/file"

const (
	defaultMaxGoroutine   = 20
	defaultCopyBufferSize = 20
)

// SynchronizerOption sets options such as FileCopier, MaxGoroutine and CopyBufferSize
type SynchronizerOption interface {
	apply(option *synchronizerOption)
}

type synchronizerOption struct {
	maxGoroutine   int
	copyBufferSize int
	fileCopier     syncFile.Copier
}

var defaultOptions = synchronizerOption{
	maxGoroutine:   defaultMaxGoroutine,
	copyBufferSize: defaultCopyBufferSize,
	fileCopier:     &syncFile.BasicCopy{},
}

type funcSynchronizerOption struct {
	f func(*synchronizerOption)
}

func (fo *funcSynchronizerOption) apply(so *synchronizerOption) {
	fo.f(so)
}

func newFuncSynchronizerOption(f func(*synchronizerOption)) *funcSynchronizerOption {
	return &funcSynchronizerOption{
		f: f,
	}
}

// MaxGoroutine lets you set the maximum goroutine that are allow for copying files.
func MaxGoroutine(m int) SynchronizerOption {
	return newFuncSynchronizerOption(func(o *synchronizerOption) {
		if m > 0 {
			o.maxGoroutine = m
		}

	})
}

// CopyBufferSize lets you set the copy channel buffer size.
func CopyBufferSize(s int) SynchronizerOption {
	return newFuncSynchronizerOption(func(o *synchronizerOption) {
		if s > 0 {
			o.copyBufferSize = s
		}
	})
}

// FileCopier lets you set up the syncFile.Copier, useful for testing purpose.
func FileCopier(fc syncFile.Copier) SynchronizerOption {
	return newFuncSynchronizerOption(func(o *synchronizerOption) {
		o.fileCopier = fc
	})
}
