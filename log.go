package bltfs

import (
	"sync"

	"hpt.space/bltfs/backend"
	pb "hpt.space/bltfs/proto"
)

type synchronizedWriter struct {
	mu struct {
		sync.Mutex
		backend backend.Interface
	}
}

func (rw *synchronizedWriter) Read(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	return rw.mu.backend.Read(p)
}

func (rw *synchronizedWriter) Write(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	return rw.mu.backend.Write(p)
}

type Log interface {
	Create(*pb.Entry)
	Remove(*pb.Entry)
	Change(*pb.Entry)
}

type Incremental struct {
	*pb.Log
}

type Differential struct {
	*pb.Log
}

func NewIncremental(epoch *Incremental) *Incremental {
	return &Incremental{
		Log: &pb.Log{
			Class:   pb.Log_INC,
			Prev:    epoch.Block,
			Entries: make([]*pb.Entry, 0),
			Extents: make([]*pb.Extent, 0),
		},
	}
}

func (inc *Incremental) Create(e *pb.Entry) {
	e.Operation = pb.Entry_ADD

	inc.Entries = append(inc.Entries, e)
}

func (inc *Incremental) Remove(e *pb.Entry) {
	e.Operation = pb.Entry_RM

	inc.Entries = append(inc.Entries, e)
}

func (inc *Incremental) Change(e *pb.Entry) {
	e.Operation = pb.Entry_CH

	inc.Entries = append(inc.Entries, e)
}

func NewDifferential(epoch *pb.Index) *Differential {
	return &Differential{
		Log: &pb.Log{
			Class:   pb.Log_DIFF,
			Prev:    epoch.Block,
			Entries: make([]*pb.Entry, 0),
			Extents: make([]*pb.Extent, 0),
		},
	}
}

func (diff *Differential) Create(e *pb.Entry) {
	e.Operation = pb.Entry_ADD

	diff.Entries = append(diff.Entries, e)
}

func (diff *Differential) Remove(e *pb.Entry) {
	e.Operation = pb.Entry_RM

	diff.Entries = append(diff.Entries, e)
}

func (diff *Differential) Change(e *pb.Entry) {
	e.Operation = pb.Entry_CH

	diff.Entries = append(diff.Entries, e)
}
