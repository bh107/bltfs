package ltfs

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"testing"

	"github.com/kr/pretty"
	"hpt.space/bltfs/util/testutil"
)

func TestLabelVolume(t *testing.T) {
	expected := "VOL1A00001L             LTFS         TEST                                      4"
	vlabel := makeLabelVolume("A00001")

	if string(vlabel[:]) != expected {
		t.Fatal("unexpected VOL1 label")
	}
}

func makeTestLabel() LabelLTFS {
	return LabelLTFS{
		XMLName:        xml.Name{Space: "", Local: "ltfslabel"},
		Version:        Version,
		Creator:        Creator,
		FormatTime:     testutil.TestTime,
		VolumeUUID:     testutil.TestUUID,
		Partition:      "b",
		IndexPartition: "a",
		DataPartition:  "b",
		BlockSize:      524288,
		Compression:    true,
	}
}

func TestLTFSLabel(t *testing.T) {
	xmlExpected := `<ltfslabel version="2.2.0">
  <creator>hpt.space bLTFS 0.0.1</creator>
  <formattime>2017-03-07T13:27:38.192689471Z</formattime>
  <volumeuuid>df925be0-44c0-4e49-af6a-7c3aa5b36d35</volumeuuid>
  <location>
    <partition>b</partition>
  </location>
  <partitions>
    <index>a</index>
    <data>b</data>
  </partitions>
  <blocksize>524288</blocksize>
  <compression>true</compression>
</ltfslabel>`

	label0 := makeTestLabel()

	buf, err := xml.MarshalIndent(label0, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	label1 := LabelLTFS{}

	if err := xml.Unmarshal(buf, &label1); err != nil {
		t.Fatal(err)
	}

	if xmlExpected != string(buf) {
		fmt.Printf("EXPECTED:\n%s\n", xmlExpected)
		fmt.Println()
		fmt.Printf("GOT:\n%s\n", string(buf))

		t.Error("xmlExpected != buf")
	}

	if !reflect.DeepEqual(label0, label1) {
		fmt.Println("EXPECTED:")
		pretty.Println(label0)
		fmt.Println("GOT:")
		pretty.Println(label1)

		t.Error("label0 != label1")
	}
}
