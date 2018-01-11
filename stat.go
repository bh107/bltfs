package bltfs

import "os"

func (b *Store) Lstat(path string) (os.FileInfo, error) {
	return b.Stat(path)
}

func (b *Store) Stat(path string) (os.FileInfo, error) {
	var es entryStat
	e, err := b.idx.stat(path)
	if err != nil {
		return nil, err
	}

	es.e = e

	return &es, nil
}
