package bltfs

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	pb "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"hpt.space/bltfs/ltfs"
	"hpt.space/bltfs/proto"
	"hpt.space/bltfs/util/xmlutil"
)

// Index is the binary index computed from the LTFS index.
type index struct {
	db      *bolt.DB
	blkSize uint64

	extents extents

	root *proto.Entry

	nextUID int

	meta struct {
		uuid    uuid.UUID
		gen     int
		prevgen int
	}
}

func NewIndex(idx *ltfs.Index, db *bolt.DB) (*index, error) {
	binIdx := &index{
		db: db,
	}

	type wrap struct {
		path string
		buf  []byte
	}

	var ws []*wrap

	collector := make(chan *wrap)
	done := make(chan struct{})

	go func() {
		for w := range collector {
			ws = append(ws, w)
		}

		close(done)
	}()

	var tmp proto.Entry
	// get the protobuf representation of the ltfs.Directory
	if err := proto.MarshalDirectory(idx.Root, &tmp); err != nil {
		panic(err)
	}

	// marshal to bytes
	buf, err := pb.Marshal(&tmp)
	if err != nil {
		panic(err)
	}

	collector <- &wrap{"/", buf}

	for _, f := range idx.Root.Contents.Files {
		// get the protobuf representation of the ltfs.Directory
		if err := proto.MarshalFile(f, &tmp); err != nil {
			panic(err)
		}

		// marshal to bytes
		buf, err := pb.Marshal(&tmp)
		if err != nil {
			panic(err)
		}

		collector <- &wrap{filepath.Join("/", f.Name), buf}
	}

	// the visit function is called concurrently, so we take care not to fuck
	// this up.
	begin := time.Now()
	idx.Root.VisitAllEntries(func(d *ltfs.Directory, subtree string) {
		var pbentry proto.Entry

		// get the protobuf representation of the ltfs.Directory
		if err := proto.MarshalDirectory(d, &pbentry); err != nil {
			panic(err)
		}

		// compose path name (insert root, add subtree, the directory we are in and
		// then directory we're inserting)
		path := filepath.Join("/", subtree, d.Name)
		path = path + "/"

		// marshal to bytes
		buf, err := pb.Marshal(&pbentry)
		if err != nil {
			panic(err)
		}

		collector <- &wrap{path, buf}

		for _, file := range d.Contents.Files {
			var pbentry proto.Entry

			// get the protobuf representation of the ltfs.File
			if err := proto.MarshalFile(file, &pbentry); err != nil {
				panic(err)
			}

			// compose path name (insert root, add subtree and then file name)
			path := filepath.Join("/", subtree, d.Name, file.Name)

			// marshal to bytes
			buf, err := pb.Marshal(&pbentry)
			if err != nil {
				panic(err)
			}

			collector <- &wrap{path, buf}
		}
	})

	close(collector)
	<-done
	fmt.Printf("building protobufs from parsed LTFS index took: %v\n", time.Since(begin))

	fmt.Printf("number of entries: %d\n", len(ws))

	// sort the index entries according to the filepath
	begin = time.Now()
	sort.Slice(ws, func(i, j int) bool {
		return ws[i].path < ws[j].path
	})
	fmt.Printf("sorting entries took: %v\n", time.Since(begin))

	// insert into database
	begin = time.Now()
	err = db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("index"))
		if bkt == nil {
			return errors.New("index bucket not found")
		}

		// We set the fill percentage to 100%. This ensures that Bolt doesn't split
		// pages before the page is full. This is good when we only expect
		// read-only on this index.
		bkt.FillPercent = 1

		// insert
		for _, w := range ws {
			if err := bkt.Put([]byte(w.path), w.buf); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	fmt.Printf("inserting to BoltDB took: %v\n", time.Since(begin))

	return binIdx, nil
}

func (idx *index) MakeLTFSIndex() (*ltfs.Index, error) {
	tree, err := idx.root.MakeTree()
	if err != nil {
		return nil, err
	}

	ltfsIndex := ltfs.Index{
		XMLName: xml.Name{Space: "", Local: "ltfsindex"},
		IndexPreface: ltfs.IndexPreface{
			Version:    ltfs.Version,
			Creator:    ltfs.Creator,
			VolumeUUID: idx.meta.uuid,
			Generation: idx.meta.gen,
			UpdateTime: xmlutil.TimeNow(),
			Partition:  "a", // TODO(kbj): DO NOT HARD CODE THIS
			StartBlock: 6,
			PreviousGeneration: ltfs.PreviousGeneration{
				Partition:  "b",
				StartBlock: 20,
			},
			HighestFileUID: 4,
		},

		Root: tree,
	}

	return &ltfsIndex, nil
}

// Insert inserts a *pb.Entry into the binary index.
func (idx *index) Insert(path string, entry *proto.Entry) error {
	buf, err := pb.Marshal(entry)
	if err != nil {
		return errors.Wrap(err, "failed to marshal entry")
	}

	return idx.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("index"))
		if bkt == nil {
			return errors.New("index bucket not found")
		}

		if err := bkt.Put([]byte(path), buf); err != nil {
			return errors.Wrap(err, "failed to insert entry")
		}

		return nil
	})
}

// Stat returns the entry at path.
func (idx *index) stat(path string) (*proto.Entry, error) {
	var entry proto.Entry

	err := idx.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("index"))
		if bkt == nil {
			return errors.New("index bucket not found")
		}

		if v := bkt.Get([]byte(path)); v != nil {
			if err := pb.Unmarshal(v, &entry); err != nil {
				return errors.Wrap(err, "failed to unmarshal entry")
			}
		}

		return errors.New("path not found")
	})

	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (idx *index) Marshal() (*proto.Entry, error) {
	// TODO(kbj): this function is a bit hairy.

	type chainT struct {
		path []byte
		dir  *proto.Directory
		prev *chainT
	}

	base := &proto.Directory{
		Entries: make([]*proto.Entry, 0),
	}

	path := []byte("/")

	chain := &chainT{
		path: path,
		dir:  base,
	}

	err := idx.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("index"))
		if bkt == nil {
			return errors.New("index bucket not found")
		}

		// iterate over all pairs in the bucket (in-order)
		err := bkt.ForEach(func(k, v []byte) error {
			var entry proto.Entry
			if err := pb.Unmarshal(v, &entry); err != nil {
				return errors.Wrap(err, "failed to unmarshal buffer")
			}

			// if we see a new prefix, move back up the chain
			for !bytes.HasPrefix(k, chain.path) {
				chain = chain.prev
			}

			if chain.dir.Entries == nil {
				// initialize the entries
				chain.dir.Entries = make([]*proto.Entry, 0)
			}

			// if the entry is a directory, add it to the chain
			switch x := entry.Elem.(type) {
			case *proto.Entry_Dir:
				// add the directory entry before updating the chain
				chain.dir.Entries = append(chain.dir.Entries, &entry)

				// update the path chain
				new := &chainT{
					path: k,
					dir:  x.Dir,
					prev: chain,
				}

				chain = new

			case *proto.Entry_File:
				chain.dir.Entries = append(chain.dir.Entries, &entry)
			}

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return base.Entries[0], nil
}

func (idx *index) Scan(path string) ([]*proto.Entry, error) {
	return idx.scan(path, true)
}

func (idx *index) List(path string) ([]*proto.Entry, error) {
	return idx.scan(path, false)
}

func (idx *index) scan(path string, resursive bool) ([]*proto.Entry, error) {
	var entries []*proto.Entry

	err := idx.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte("index"))
		if bkt == nil {
			return errors.New("index bucket not found")
		}

		c := bkt.Cursor()

		prefix := []byte(path)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if !resursive {
				k = k[len(prefix):]
				if bytes.IndexByte(k, '/') != -1 {
					continue
				}
			}

			var entry proto.Entry

			// unmarshal the entry
			if err := pb.Unmarshal(v, &entry); err != nil {
				return errors.Wrap(err, "failed to unmarshal entry")
			}

			entries = append(entries, &entry)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return entries, nil
}

type extentList []*proto.Extent

func (es extentList) last() *proto.Extent {
	return es[len(es)-1]
}

type extents map[uint64]extentList

func (e extents) last(f *File) *proto.Extent {
	if es, ok := e[f.id]; ok {
		return es.last()
	}

	return nil
}

func (idx *index) addExtent(f *File, e *proto.Extent) {
	prev := idx.extents.last(f)
	if prev.Block+prev.Length/idx.blkSize == e.Block {

	}
}

type indexWriter struct {
	idx *index
	rw  io.ReadWriter
}

func (idx *index) newIndexWriter() *indexWriter {
	return &indexWriter{
		idx: idx,
	}
}

func (w *indexWriter) start(ctx context.Context, pol *RecoveryPolicy) {
	go func() {
		for {
			select {
			case <-time.After(pol.FullIndexInterval):
			}
		}
	}()
}

// ReadLTFSIndex reads the LTFS index from the device and returns a ltfs.Index
// representation.
func (b *Store) ReadLTFSIndex() (*ltfs.Index, error) {
	// seek to EOD
	if err := b.mu.backend.Locate(0, TapeBlockMax); err != nil {
		return nil, errors.Wrap(err, "failed to seek to EOD")
	}

	// space backwards to find the LTFS index
	if err := b.mu.backend.SpaceFMB(2); err != nil {
		return nil, errors.Wrap(err, "failed to space backward")
	}

	buf, err := b.ReadFile()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read file")
	}

	var idx ltfs.Index

	// unmarshal the LTFS index
	if err := xml.Unmarshal(buf, &idx); err != nil {
		return nil, err
	}

	return &idx, nil
}
