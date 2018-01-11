package ltfs

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"hpt.space/bltfs/util/testutil"
)

func TestFile(t *testing.T) {
	xmlExpected := `<file>
  <fileuid>4</fileuid>
  <name>testfile.txt</name>
  <length>10</length>
  <creationtime>2017-03-07T13:27:38.192689471Z</creationtime>
  <changetime>2017-03-07T13:27:38.192689471Z</changetime>
  <modifytime>2017-03-07T13:27:38.192689471Z</modifytime>
  <accesstime>2017-03-07T13:27:38.192689471Z</accesstime>
  <backuptime>2017-03-07T13:27:38.192689471Z</backuptime>
  <readonly>true</readonly>
  <extentinfo>
    <extent>
      <partition>a</partition>
      <startblock>4</startblock>
      <byteoffset>0</byteoffset>
      <bytecount>5</bytecount>
      <fileoffset>0</fileoffset>
    </extent>
    <extent>
      <partition>a</partition>
      <startblock>9</startblock>
      <byteoffset>0</byteoffset>
      <bytecount>5</bytecount>
      <fileoffset>5</fileoffset>
    </extent>
  </extentinfo>
</file>`

	file0 := File{
		XMLName:            xml.Name{Space: "", Local: "file"},
		FileUID:            4,
		Name:               "testfile.txt",
		Length:             10,
		CreationTime:       testutil.TestTime,
		ChangeTime:         testutil.TestTime,
		ModifyTime:         testutil.TestTime,
		AccessTime:         testutil.TestTime,
		BackupTime:         testutil.TestTime,
		ReadOnly:           true,
		ExtendedAttributes: nil,
		ExtentInfo: []*Extent{
			&Extent{
				Partition:  "a",
				StartBlock: 4,
				ByteOffset: 0,
				ByteCount:  5,
				FileOffset: 0,
			},
			&Extent{
				Partition:  "a",
				StartBlock: 9,
				ByteOffset: 0,
				ByteCount:  5,
				FileOffset: 5,
			},
		},
	}

	buf, err := xml.MarshalIndent(file0, "", "  ")
	if err != nil {
		t.Error(err)
	}

	if xmlExpected != string(buf) {
		fmt.Printf("EXPECTED:\n%s\n", xmlExpected)
		fmt.Println()
		fmt.Printf("GOT:\n%s\n", string(buf))

		t.Error("xmlExpected != buf")
	}

	file1 := File{}

	if err := xml.Unmarshal(buf, &file1); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(file0, file1) {
		fmt.Println("EXPECTED:")
		pretty.Println(file0)
		fmt.Println("GOT:")
		pretty.Println(file1)

		t.Error("file0 != file1")
	}
}
