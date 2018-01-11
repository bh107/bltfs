package bltfs

import "time"

type RecoveryPolicy struct {
	FullIndexInterval time.Duration
	DifferentialAfter uint64
	IncrementalAfter  uint64
}

var DefaultRecoveryPolicy = RecoveryPolicy{
	FullIndexInterval: 1 * time.Hour,
}

type storeOptions struct {
	blkSize   uint64
	pol       RecoveryPolicy
	reporter  Reporter
	filedebug bool
}

type StoreOption func(*storeOptions)

func WithFileDebug() StoreOption {
	return func(o *storeOptions) {
		o.filedebug = true
	}
}

func WithRecoveryPolicy(pol RecoveryPolicy) StoreOption {
	return func(o *storeOptions) {
		o.pol = pol
	}
}

func WithBlockSize(blkSize uint64) StoreOption {
	return func(o *storeOptions) {
		o.blkSize = blkSize
	}
}

func WithReporter(reporter Reporter) StoreOption {
	return func(o *storeOptions) {
		o.reporter = reporter
	}
}
