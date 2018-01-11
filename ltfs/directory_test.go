package ltfs

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"hpt.space/bltfs/util/testutil"
)

func makeTestDirectory() *Directory {
	return &Directory{
		XMLName:      xml.Name{Space: "", Local: "directory"},
		FileUID:      1,
		Name:         "LTFS Volume Name",
		CreationTime: testutil.TestTime,
		ChangeTime:   testutil.TestTime,
		ModifyTime:   testutil.TestTime,
		AccessTime:   testutil.TestTime,
		BackupTime:   testutil.TestTime,
		Contents: &Contents{
			Directories: []*Directory{
				&Directory{
					XMLName:      xml.Name{Space: "", Local: "directory"},
					FileUID:      2,
					Name:         "directory1",
					CreationTime: testutil.TestTime,
					ChangeTime:   testutil.TestTime,
					ModifyTime:   testutil.TestTime,
					AccessTime:   testutil.TestTime,
					BackupTime:   testutil.TestTime,
					ReadOnly:     false,
					Contents: &Contents{
						Directories: []*Directory{
							&Directory{
								XMLName:      xml.Name{Space: "", Local: "directory"},
								FileUID:      3,
								Name:         "subdir1",
								CreationTime: testutil.TestTime,
								ChangeTime:   testutil.TestTime,
								ModifyTime:   testutil.TestTime,
								AccessTime:   testutil.TestTime,
								BackupTime:   testutil.TestTime,
								ReadOnly:     false,
							},
						},
					},
				},
			},
			Files: []*File{
				&File{
					XMLName:            xml.Name{Space: "", Local: "file"},
					FileUID:            4,
					Name:               "testfile.txt",
					Length:             5,
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
					},
				},
			},
		},
	}
}

func TestDirectory(t *testing.T) {
	xmlExpected := `<directory>
  <fileuid>1</fileuid>
  <name>LTFS Volume Name</name>
  <creationtime>2017-03-07T13:27:38.192689471Z</creationtime>
  <changetime>2017-03-07T13:27:38.192689471Z</changetime>
  <modifytime>2017-03-07T13:27:38.192689471Z</modifytime>
  <accesstime>2017-03-07T13:27:38.192689471Z</accesstime>
  <backuptime>2017-03-07T13:27:38.192689471Z</backuptime>
  <contents>
    <file>
      <fileuid>4</fileuid>
      <name>testfile.txt</name>
      <length>5</length>
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
      </extentinfo>
    </file>
    <directory>
      <fileuid>2</fileuid>
      <name>directory1</name>
      <creationtime>2017-03-07T13:27:38.192689471Z</creationtime>
      <changetime>2017-03-07T13:27:38.192689471Z</changetime>
      <modifytime>2017-03-07T13:27:38.192689471Z</modifytime>
      <accesstime>2017-03-07T13:27:38.192689471Z</accesstime>
      <backuptime>2017-03-07T13:27:38.192689471Z</backuptime>
      <contents>
        <directory>
          <fileuid>3</fileuid>
          <name>subdir1</name>
          <creationtime>2017-03-07T13:27:38.192689471Z</creationtime>
          <changetime>2017-03-07T13:27:38.192689471Z</changetime>
          <modifytime>2017-03-07T13:27:38.192689471Z</modifytime>
          <accesstime>2017-03-07T13:27:38.192689471Z</accesstime>
          <backuptime>2017-03-07T13:27:38.192689471Z</backuptime>
        </directory>
      </contents>
    </directory>
  </contents>
</directory>`

	dir0 := makeTestDirectory()

	buf, err := xml.MarshalIndent(dir0, "", "  ")
	if err != nil {
		t.Error(err)
	}

	dir1 := Directory{}

	if err := xml.Unmarshal(buf, &dir1); err != nil {
		t.Error(err)
	}

	if xmlExpected != string(buf) {
		fmt.Printf("EXPECTED:\n%s\n", xmlExpected)
		fmt.Println()
		fmt.Printf("GOT:\n%s\n", string(buf))

		t.Error("xmlExpected != buf")
	}

	if !reflect.DeepEqual(dir0, &dir1) {
		fmt.Println("EXPECTED:")
		pretty.Println(dir0)
		fmt.Println("GOT:")
		pretty.Println(dir1)

		t.Error("dir0 != dir1")
	}
}
