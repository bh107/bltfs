package xmlutil_test

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"testing"

	"github.com/kr/pretty"

	"hpt.space/bltfs/util/testutil"
	"hpt.space/bltfs/util/xmlutil"
)

func TestXMLTime(t *testing.T) {
	xmlExpected := `<Time>2017-03-07T13:27:38.192689471Z</Time>`

	t0 := testutil.TestTime

	buf, err := xml.Marshal(t0)
	if err != nil {
		t.Error(err)
	}

	if xmlExpected != string(buf) {
		fmt.Printf("EXPECTED:\n%s\n", xmlExpected)
		fmt.Println()
		fmt.Printf("GOT:\n%s\n", string(buf))

		t.Error("xmlExpected != buf")
	}

	t1 := xmlutil.Time{}

	if err := xml.Unmarshal(buf, &t1); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(t0, t1) {
		fmt.Println("EXPECTED:")
		pretty.Println(t0)
		fmt.Println("GOT:")
		pretty.Println(t1)

		t.Error("t0 != t1")
	}

}
