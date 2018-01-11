package bltfs

import (
	"context"
	"errors"
	"sync"

	"hpt.space/bltfs/backend"
	"hpt.space/bltfs/ltfs"
)

const (
	// TapeBlockMax is the higest addressable tape block.
	TapeBlockMax = 0xFFFFFFFFFFFFFFFF
)

var (
	// ErrIO is an I/O error.
	ErrIO = errors.New("I/O error")

	// ErrEOD is an End-Of-Device error.
	ErrEOD = errors.New("EOD")

	// ErrBOT is an Beginning-Of-Tape error.
	ErrBOT = errors.New("BOT")

	// ErrNotReady signifies that the device was not ready.
	ErrNotReady = errors.New("Device not ready")
)

// Store is a bLTFS store.
type Store struct {
	idx *index
	rw  *synchronizedWriter

	mu struct {
		sync.Mutex
		backend backend.Interface
	}

	ltfs struct {
		curr *ltfs.Index
		prev *ltfs.Index
	}

	sopts storeOptions
}

// Open opens a new bLTFS store using the given backend.
func Open(backend backend.Interface, opts ...StoreOption) (*Store, error) {
	return OpenContext(context.Background(), backend, opts...)
}

// OpenContext opens a new bLTFS store using the given backend and
// the given context.
func OpenContext(ctx context.Context, backend backend.Interface, opts ...StoreOption) (*Store, error) {
	s := &Store{}

	s.mu.backend = backend
	s.rw = &synchronizedWriter{}
	s.rw.mu.backend = backend

	s.sopts.blkSize = 512 * 1024
	s.sopts.pol = DefaultRecoveryPolicy

	for _, opt := range opts {
		opt(&s.sopts)
	}

	// initialize
	if err := backend.Load(); err != nil {
		return nil, err
	}

	// seek to EOD on data partition
	if err := s.mu.backend.Locate(1, TapeBlockMax); err != nil {
		return nil, err
	}

	// set active partition
	if err := s.mu.backend.SetPartition(1); err != nil {
		return nil, err
	}

	return s, nil
}

// Close closes the bLTFS store.
func (s *Store) Close() error {
	return nil
}

/*
// Recover recovers a store.
func (s *Store) Recover() error {
	// seek to EOD on data partition
	if err := s.mu.backend.Locate(1, TapeBlockMax); err != nil {
		return errors.Wrap(err, "failed to seek to EOD")
	}

	// Recovery
	//
	// We expect that no LTFS index has been written, so we space backward one
	// file mark at a time.
	if err := s.mu.backend.SpaceFMB(1); err != nil {
		return err
	}

	for {
		buf, err := ioutil.ReadAll(s.mu.backend)
		if err != nil {
			return err
		}

		var pblog pb.Log
		if err := proto.Unmarshal(buf, &pblog); err != nil {
			return err
		}

		switch pblog.Class {
		case pb.Log_INC:
			// seek to the previous incremental
			if err := s.mu.backend.Locate(1, pblog.Prev); err != nil {
				return err
			}

		case pb.Log_DIFF:

		}
	}
}
*/
