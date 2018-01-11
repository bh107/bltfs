package ltfs

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"hpt.space/bltfs/util/testutil"
)

func makeTestIndex() Index {
	return Index{
		XMLName: xml.Name{Space: "", Local: "ltfsindex"},
		IndexPreface: IndexPreface{
			Version:    Version,
			Creator:    Creator,
			VolumeUUID: testutil.TestUUID,
			Generation: 3,
			Comment:    "A sample LTFS index",
			UpdateTime: testutil.TestTime,
			Partition:  "a",
			StartBlock: 6,
			PreviousGeneration: PreviousGeneration{
				Partition:  "b",
				StartBlock: 20,
			},
			AllowPolicyUpdate: true,
			DataPlacementPolicy: DataPlacementPolicy{
				Size: 1048576,
				Name: []string{"*.txt", "*.bin"},
			},
			HighestFileUID: 4,
		},
		Root: makeTestDirectory(),
	}
}

func TestLTFSIndex(t *testing.T) {
	xmlExpected := `<ltfsindex version="2.2.0">
  <creator>hpt.space bLTFS 0.0.1</creator>
  <volumeuuid>df925be0-44c0-4e49-af6a-7c3aa5b36d35</volumeuuid>
  <generationnumber>3</generationnumber>
  <comment>A sample LTFS index</comment>
  <updatetime>2017-03-07T13:27:38.192689471Z</updatetime>
  <location>
    <partition>a</partition>
    <startblock>6</startblock>
  </location>
  <previousgenerationlocation>
    <partition>b</partition>
    <startblock>20</startblock>
  </previousgenerationlocation>
  <allowpolicyupdate>true</allowpolicyupdate>
  <dataplacementpolicy>
    <indexpartitioncriteria>
      <size>1048576</size>
      <name>*.txt</name>
      <name>*.bin</name>
    </indexpartitioncriteria>
  </dataplacementpolicy>
  <highestfileuid>4</highestfileuid>
  <directory>
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
  </directory>
</ltfsindex>`

	idx0 := makeTestIndex()

	buf, err := xml.MarshalIndent(idx0, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if xmlExpected != string(buf) {
		fmt.Printf("EXPECTED:\n%s\n", xmlExpected)
		fmt.Println()
		fmt.Printf("GOT:\n%s\n", string(buf))

		t.Error("xmlExpected != buf")
	}

	idx1 := Index{}

	if err := xml.Unmarshal(buf, &idx1); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(idx0, idx1) {
		fmt.Println("EXPECTED:")
		pretty.Println(idx0)
		fmt.Println("GOT:")
		pretty.Println(idx1)

		t.Error("idx0 != idx1")
	}
}
