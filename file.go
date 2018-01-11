package bltfs

import (
	"os"
	"time"

	"hpt.space/bltfs/proto"
)

type entryStat struct {
	e *proto.Entry
}

// Size is part of the os.FileInfo interface
func (es *entryStat) Size() int64 {
	if x, ok := es.e.Elem.(*proto.Entry_File); ok {
		return int64(x.File.Length)
	}

	panic("Size() not implemented for directories")
}

// IsDir is part of the os.FileInfo interface
func (es *entryStat) IsDir() bool {
	_, ok := es.e.Elem.(*proto.Entry_Dir)

	return ok
}

func (es *entryStat) Mode() os.FileMode  { return os.ModePerm }
func (es *entryStat) ModTime() time.Time { return time.Unix(0, es.e.ModifyTime) }
func (es *entryStat) Sys() interface{}   { return es.e }
func (es *entryStat) Name() string       { return es.e.Name }

type fileOptions struct {
	noBatch bool
}

type FileOption func(*fileOptions)

func WithNoBatch() FileOption {
	return func(o *fileOptions) {
		o.noBatch = true
	}
}

type File struct {
	id uint64

	rw   *synchronizedWriter
	idx  *index
	path string

	fopts fileOptions
}

func (f *File) String() string {
	return f.path
}

func (s *Store) Create(path string, opts ...FileOption) (*File, error) {
	return s.Open(path, opts...)
}

func (s *Store) Open(path string, opts ...FileOption) (*File, error) {
	f := &File{
		rw:   s.rw,
		idx:  s.idx,
		path: path,
	}

	for _, opt := range opts {
		opt(&f.fopts)
	}

	return f, nil
}

func (f *File) Name() string {
	return f.path
}

func (f *File) Close() error {
	return nil
}

func (f *File) Write(p []byte) (n int, err error) {
	return f.rw.Write(p)
}

func (f *File) Read(p []byte) (n int, err error) {
	return f.rw.Read(p)
}

func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	panic("not implemented")
}
