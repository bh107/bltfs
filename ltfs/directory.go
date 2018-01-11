package ltfs

import (
	"encoding/xml"
	"path/filepath"
	"sync"

	"hpt.space/bltfs/util/xmlutil"
)

// Directory represents an LTFS directory construct.
type Directory struct {
	XMLName      xml.Name     `xml:"directory"`
	FileUID      int          `xml:"fileuid"`
	Name         string       `xml:"name"`
	CreationTime xmlutil.Time `xml:"creationtime"`
	ChangeTime   xmlutil.Time `xml:"changetime"`
	ModifyTime   xmlutil.Time `xml:"modifytime"`
	AccessTime   xmlutil.Time `xml:"accesstime"`
	BackupTime   xmlutil.Time `xml:"backuptime"`
	ReadOnly     bool         `xml:"readonly,omitempty"`
	Contents     *Contents    `xml:"contents"`
}

// Contents is the structure that contains the Files and Directories of a given
// directory.
type Contents struct {
	Files       []*File      `xml:"file"`
	Directories []*Directory `xml:"directory"`
}

func (d *Directory) visitAll(fn func(*Directory, string), subtree string) {
	var wg sync.WaitGroup

	for _, dir := range d.Contents.Directories {
		wg.Add(1)

		go func(d *Directory) {
			defer wg.Done()

			fn(d, subtree)

			d.visitAll(fn, filepath.Join(subtree, d.Name))
		}(dir)
	}

	wg.Wait()
}

// VisitAllEntries visits all entries.
func (d *Directory) VisitAllEntries(fn func(d *Directory, subtree string)) {
	d.visitAll(fn, "")
}
