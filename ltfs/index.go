package ltfs

import (
	"encoding/xml"
	"io/ioutil"

	"github.com/google/uuid"
	"hpt.space/bltfs/util/xmlutil"
)

// Index represents the LTFS index construct.
type Index struct {
	XMLName xml.Name `xml:"ltfsindex"`
	IndexPreface
	Root *Directory `xml:"directory"`
}

// IndexPreface contains metadata for the index construct.
type IndexPreface struct {
	Version    string       `xml:"version,attr"`
	Creator    string       `xml:"creator"`
	VolumeUUID uuid.UUID    `xml:"volumeuuid"`
	Generation int          `xml:"generationnumber"`
	Comment    string       `xml:"comment"`
	UpdateTime xmlutil.Time `xml:"updatetime"`
	Partition  string       `xml:"location>partition"`
	StartBlock int          `xml:"location>startblock"`
	PreviousGeneration
	AllowPolicyUpdate bool `xml:"allowpolicyupdate"`
	DataPlacementPolicy
	HighestFileUID int `xml:"highestfileuid"`
}

// DataPlacementPolicy represents the LTFS dataplacementpolicy tag.
type DataPlacementPolicy struct {
	Size int      `xml:"dataplacementpolicy>indexpartitioncriteria>size"`
	Name []string `xml:"dataplacementpolicy>indexpartitioncriteria>name"`
}

// PreviousGeneration is an LTFS tag.
type PreviousGeneration struct {
	Partition  string `xml:"previousgenerationlocation>partition"`
	StartBlock int    `xml:"previousgenerationlocation>startblock"`
}

// LoadIndexFromFile loads an LTFS index from a file.
func LoadIndexFromFile(path string) (*Index, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var idx Index

	// unmarshal the LTFS index
	if err := xml.Unmarshal(buf, &idx); err != nil {
		return nil, err
	}

	return &idx, nil
}
