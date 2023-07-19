package directory

import syncFile "gosync/pkg/file"

type synchronizerOption struct {
	maxGoroutine   int
	copyBufferSize int
	fileCopier     syncFile.Copier
}

const (
	defaultMaxGoroutine   = 20
	defaultCopyBufferSize = 20
)

var defaultOptions = synchronizerOption{
	maxGoroutine:   defaultMaxGoroutine,
	copyBufferSize: defaultCopyBufferSize,
	fileCopier:     syncFile.BasicCopy{},
}

type SynchronizerOption interface {
	apply(option *synchronizerOption)
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

func MaxGoroutine(m int) SynchronizerOption {
	return newFuncSynchronizerOption(func(o *synchronizerOption) {
		if m > 0 {
			o.maxGoroutine = m
		}

	})
}

func copyBufferSize(s int) SynchronizerOption {
	return newFuncSynchronizerOption(func(o *synchronizerOption) {
		if s > 0 {
			o.copyBufferSize = s
		}
	})
}

func FileCopier(fc syncFile.Copier) SynchronizerOption {
	return newFuncSynchronizerOption(func(o *synchronizerOption) {
		o.fileCopier = fc
	})
}
