package ltfs

import (
	"bytes"
	"encoding/xml"
	"strings"

	"github.com/google/uuid"
	"hpt.space/bltfs/util/xmlutil"
)

// LabelVolume is the ANSI volume label.
type LabelVolume [80]byte

func makeLabelVolume(serial string) LabelVolume {
	if len(serial) != 6 {
		panic("serial MUST be have lenght 6")
	}

	var label LabelVolume
	var buf bytes.Buffer

	// label identifier and number
	buf.WriteString("VOL1")

	// volume identifier
	buf.WriteString(serial)

	// volume accessibility
	buf.WriteString("L")

	// reserved
	buf.WriteString(strings.Repeat(" ", 13))

	// implementation identifier
	buf.WriteString("LTFS")
	buf.WriteString(strings.Repeat(" ", 13-len("LTFS")))

	// owner identifier
	buf.WriteString("TEST")
	buf.WriteString(strings.Repeat(" ", 14-len("TEST")))

	// reserved
	buf.WriteString(strings.Repeat(" ", 28))

	// label standard version
	buf.WriteString("4")

	copy(label[:], buf.Bytes())

	return label
}

// LabelLTFS represents the LTFS label construct.
type LabelLTFS struct {
	XMLName        xml.Name     `xml:"ltfslabel"`
	Version        string       `xml:"version,attr"`
	Creator        string       `xml:"creator"`
	FormatTime     xmlutil.Time `xml:"formattime"`
	VolumeUUID     uuid.UUID    `xml:"volumeuuid"`
	Partition      string       `xml:"location>partition"`
	IndexPartition string       `xml:"partitions>index"`
	DataPartition  string       `xml:"partitions>data"`
	BlockSize      int          `xml:"blocksize"`
	Compression    bool         `xml:"compression"`
}
