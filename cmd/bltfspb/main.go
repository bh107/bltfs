package main

import (
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	pb "github.com/tapebit/bltfs/protobuf"
)

func pbsize(m proto.Marshaler) int {
	buf, err := m.Marshal()
	if err != nil {
		panic(err)
	}

	return len(buf)
}

func main() {
	pbextent := pb.Extent{
		Partition: 1,
		Block:     1,
		Length:    1,
		Boffset:   0,
		Offset:    0,
	}

	pbfile := pb.Entry{
		Id:         1,
		Name:       "x",
		Readonly:   true,
		CreateTime: time.Now().UnixNano(),
		ChangeTime: time.Now().UnixNano(),
		ModifyTime: time.Now().UnixNano(),
		AccessTime: time.Now().UnixNano(),
		BackupTime: time.Now().UnixNano(),

		Elem: &pb.Entry_File{
			File: &pb.File{
				Length: 1024 * 1024,
				Extents: []*pb.Extent{
					&pbextent,
				},
			},
		},
	}

	pbemptyfile := pb.Entry{
		Id:         1,
		Name:       "x",
		Readonly:   true,
		CreateTime: time.Now().UnixNano(),
		ChangeTime: time.Now().UnixNano(),
		ModifyTime: time.Now().UnixNano(),
		AccessTime: time.Now().UnixNano(),
		BackupTime: time.Now().UnixNano(),

		Elem: &pb.Entry_File{
			File: nil,
		},
	}

	pbdir := pb.Entry{
		Id:         1,
		Name:       "x",
		Readonly:   true,
		CreateTime: time.Now().UnixNano(),
		ChangeTime: time.Now().UnixNano(),
		ModifyTime: time.Now().UnixNano(),
		AccessTime: time.Now().UnixNano(),
		BackupTime: time.Now().UnixNano(),

		Elem: &pb.Entry_Dir{
			Dir: nil, //&pb.Directory{},
		},
	}

	pblog := pb.Log{
		Class: pb.Log_INC,
		Prev:  1,
		Block: 1,
	}

	fmt.Printf("size of empty pb.Extent: %d\n", pbsize(&pbextent))
	fmt.Printf("size of empty pb.File (with 1 extent): %d\n", pbsize(&pbfile))
	fmt.Printf("size of empty pb.File (empty): %d\n", pbsize(&pbemptyfile))
	fmt.Printf("size of empty pb.Directory: %d\n", pbsize(&pbdir))
	fmt.Printf("size of empty pb.Log: %d\n", pbsize(&pblog))

	db, err := bolt.Open("testidx.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	buf, err := pbfile.Marshal()
	if err != nil {
		log.Fatal(err)
	}

	// insert into database
	begin := time.Now()
	for j := 0; j < 10; j++ {
		err = db.Update(func(tx *bolt.Tx) error {
			bkt, err := tx.CreateBucketIfNotExists([]byte("index"))
			if err != nil {
				return err
			}

			// We set the fill percentage to 100%. This ensures that Bolt doesn't split
			// pages before the page is full. This is good when we only expect
			// read-only on this index.
			bkt.FillPercent = 1

			// insert
			for i := 0; i < 10000; i++ {
				if err := bkt.Put([]byte(fmt.Sprintf("/%d", i)), buf); err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("inserting took: %v\n", time.Since(begin))

}
