package bltfs_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	pb "github.com/golang/protobuf/proto"
	"github.com/kr/pretty"
	"github.com/pkg/errors"

	"hpt.space/bltfs"
	"hpt.space/bltfs/ltfs"
	"hpt.space/bltfs/proto"
)

const (
	testDirectory = "/tmp/ltfs/tape"
	fixtures      = "./fixtures"
)

func setupCleanTape() string {
	dir, err := ioutil.TempDir("", "bltfstest")
	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	return dir
}

func cleanup(path string) {
	if err := os.RemoveAll(path); err != nil {
		panic(err)
	}
}

func TestLargeBinaryIndex(t *testing.T) {
	db, err := bolt.Open("idx.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("index"))
		if err != nil {
			return errors.Wrap(err, "failed to create bucket")
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	begin := time.Now()
	idx, err := ltfs.LoadIndexFromFile("./fixtures/large.schema")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("loading and parsing XML index from disk took: %v\n", time.Since(begin))

	binIdx, err := bltfs.NewIndex(idx, db)
	if err != nil {
		t.Fatal(err)
	}

	begin = time.Now()

	//dir := "/+C=DK+ST=NA+L=NA+O=NBI+OU=NA+CN=Jonas_Bardino+emailAddress=bardino@nbi.ku.dk/archive-fGAY1G/"
	dir := "/"
	begin = time.Now()
	lst, err := binIdx.Scan(dir)
	fmt.Printf("full scan took: %v\n", time.Since(begin))

	if err != nil {
		t.Fatal(err)
	}

	var nDirs, nFiles int
	for _, e := range lst {
		switch e.Elem.(type) {
		case *proto.Entry_Dir:
			nDirs++
		case *proto.Entry_File:
			nFiles++
		}
	}

	fmt.Printf("number of directories: %d\n", nDirs)
	fmt.Printf("number of files: %d\n", nFiles)

	var root proto.Entry

	begin = time.Now()
	if err := proto.MarshalDirectoryRecursive(idx.Root, &root); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("full protobuf tree building from ltfs.Index took: %v\n", time.Since(begin))

	//pretty.Println(root)
	//printTree(&root, "/")
	//printDir(&root)

	begin = time.Now()
	buf, err := pb.Marshal(&root)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("full protobuf tree marshalling took: %v\n", time.Since(begin))

	if err := ioutil.WriteFile("testpb", buf, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	begin = time.Now()
	pbdir, err := binIdx.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("full protobuf tree building from boltdb took: %v\n", time.Since(begin))

	//pretty.Println(pbdir)
	//printTree(pbdir, "/")
	//printDir(pbdir)

	begin = time.Now()
	buf, err = pb.Marshal(pbdir)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("full protobuf tree marshalling took: %v\n", time.Since(begin))

	if err := ioutil.WriteFile("testpb2", buf, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	begin = time.Now()
	buf, err = ioutil.ReadFile("testpb2")
	if err != nil {
		t.Fatal(err)
	}

	var root2 proto.Entry
	if err := pb.Unmarshal(buf, &root2); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("loading and unmarshalling full protobuf tree from disk took: %v\n", time.Since(begin))

	if !reflect.DeepEqual(pbdir, &root2) {
		t.Fatal("written and read protobuf not equal!")
	}
}

func TestSpeed(t *testing.T) {
	begin := time.Now()
	buf, err := ioutil.ReadFile("testpb2")
	if err != nil {
		t.Fatal(err)
	}

	var root proto.Entry
	if err := pb.Unmarshal(buf, &root); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("loading and unmarshalling full protobuf tree from disk took: %v\n", time.Since(begin))
}

func printTree(root *proto.Entry, subdir string) {
	fmt.Println(filepath.Join(subdir, root.Name))
	pretty.Println(root)
	switch x := root.Elem.(type) {
	case *proto.Entry_Dir:
		for _, e := range x.Dir.Entries {
			printTree(e, filepath.Join(subdir, root.Name))
		}
	}
}

func printDir(d *proto.Entry) {
	for _, e := range d.Elem.(*proto.Entry_Dir).Dir.Entries {
		fmt.Println(e.Name)
	}
}

/*
func TestBinaryIndex(t *testing.T) {
	db, err := bolt.Open("idx.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("index"))
		if err != nil {
			return errors.Wrap(err, "failed to create bucket")
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	backend, err := filedebug.Open(filepath.Join(fixtures, "large-filled-tape"))
	if err != nil {
		t.Fatal(err)
	}

	store, err := bltfs.Open(backend,
		bltfs.WithFileDebug(),
	)

	var root pb.Entry

	begin := time.Now()
	if err := pb.MarshalDirectoryRecursive(store.idx.Root, &root); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("full protobuf tree building took: %v\n", time.Since(begin))

	pretty.Println(root)

	begin = time.Now()
	buf, err := root.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("full protobuf tree marshalling took: %v\n", time.Since(begin))

	if err := ioutil.WriteFile("testpb", buf, os.ModePerm); err != nil {
		t.Fatal(err)
	}

	pbdir, err := binIdx.Build()
	if err != nil {
		t.Fatal(err)
	}

	pretty.Println(pbdir)

	buf, err = pbdir.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile("testpb2", buf, os.ModePerm); err != nil {
		t.Fatal(err)
	}
}
*/
