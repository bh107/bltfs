package ltfs

import (
	"encoding/xml"

	"hpt.space/bltfs/util/xmlutil"
)

// File represents an LTFS file construct.
type File struct {
	XMLName            xml.Name             `xml:"file"`
	FileUID            int                  `xml:"fileuid"`
	Name               string               `xml:"name"`
	Length             int                  `xml:"length"`
	CreationTime       xmlutil.Time         `xml:"creationtime"`
	ChangeTime         xmlutil.Time         `xml:"changetime"`
	ModifyTime         xmlutil.Time         `xml:"modifytime"`
	AccessTime         xmlutil.Time         `xml:"accesstime"`
	BackupTime         xmlutil.Time         `xml:"backuptime"`
	ReadOnly           bool                 `xml:"readonly"`
	ExtendedAttributes []*ExtendedAttribute `xml:"extendedattributes"`
	ExtentInfo         []*Extent            `xml:"extentinfo>extent"`
}

// ExtendedAttribute represents files extended attributes. Currently not
// implemented/supported.
type ExtendedAttribute struct{}

// Extent represents an LTFS extent construct.
type Extent struct {
	Partition  string `xml:"partition"`
	StartBlock int    `xml:"startblock"`
	ByteOffset int    `xml:"byteoffset"`
	ByteCount  int    `xml:"bytecount"`
	FileOffset int    `xml:"fileoffset"`
}
