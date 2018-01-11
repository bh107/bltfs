package proto

import (
	"encoding/xml"
	"errors"

	"hpt.space/bltfs/ltfs"
	"hpt.space/bltfs/util/xmlutil"
)

// To regenerate the protocol buffer output for this package, run
//	 go generate

//go:generate protoc bltfs.proto --go_out=.

func (e *Entry) MakeTree() (*ltfs.Directory, error) {
	d := &ltfs.Directory{
		XMLName:      xml.Name{Space: "", Local: "directory"},
		FileUID:      int(e.Id),
		Name:         e.Name,
		CreationTime: xmlutil.Unix(0, e.CreateTime),
		ChangeTime:   xmlutil.Unix(0, e.ChangeTime),
		ModifyTime:   xmlutil.Unix(0, e.ModifyTime),
		AccessTime:   xmlutil.Unix(0, e.AccessTime),
		BackupTime:   xmlutil.Unix(0, e.BackupTime),

		Contents: &ltfs.Contents{
			Directories: make([]*ltfs.Directory, 0),
			Files:       make([]*ltfs.File, 0),
		},
	}

	elem, ok := e.Elem.(*Entry_Dir)
	if !ok {
		return nil, errors.New("entry is not a directory")
	}

	for _, dentry := range elem.Dir.Entries {
		switch dentry.Elem.(type) {
		case *Entry_Dir:
			dir, err := dentry.MakeTree()
			if err != nil {
				return nil, err
			}

			d.Contents.Directories = append(d.Contents.Directories, dir)
		case *Entry_File:
			file, err := dentry.MakeFile()
			if err != nil {
				return nil, err
			}

			d.Contents.Files = append(d.Contents.Files, file)
		}
	}

	return d, nil
}

func (e *Entry) MakeFile() (*ltfs.File, error) {
	f := &ltfs.File{
		XMLName:      xml.Name{Space: "", Local: "file"},
		FileUID:      int(e.Id),
		Name:         e.Name,
		CreationTime: xmlutil.Unix(0, e.CreateTime),
		ChangeTime:   xmlutil.Unix(0, e.ChangeTime),
		ModifyTime:   xmlutil.Unix(0, e.ModifyTime),
		AccessTime:   xmlutil.Unix(0, e.AccessTime),
		BackupTime:   xmlutil.Unix(0, e.BackupTime),
	}

	elem, ok := e.Elem.(*Entry_File)
	if !ok {
		return nil, errors.New("entry is not a file")
	}

	f.Length = int(elem.File.Length)
	f.ExtentInfo = make([]*ltfs.Extent, 0)

	for _, extent := range elem.File.Extents {
		ex := extent.MakeExtent()
		f.ExtentInfo = append(f.ExtentInfo, ex)
	}

	return f, nil
}

func (e *Extent) MakeExtent() *ltfs.Extent {
	return &ltfs.Extent{
		Partition:  "b", // TODO(kbj): DO NOT HARD CODE THIS
		StartBlock: int(e.Block),
		ByteCount:  int(e.Length),
		ByteOffset: int(e.Boffset),
		FileOffset: int(e.Offset),
	}
}

func MarshalDirectoryRecursive(d *ltfs.Directory, e *Entry) error {
	if err := MarshalDirectory(d, e); err != nil {
		return err
	}

	entries := make([]*Entry, 0)

	for _, file := range d.Contents.Files {
		var pbfile Entry
		if err := MarshalFile(file, &pbfile); err != nil {
			return err
		}

		entries = append(entries, &pbfile)
	}

	for _, dir := range d.Contents.Directories {
		var pbdir Entry
		if err := MarshalDirectoryRecursive(dir, &pbdir); err != nil {
			return err
		}

		entries = append(entries, &pbdir)
	}

	// save the entries
	tmp := e.GetDir()
	tmp.Entries = entries

	return nil
}

// MarshalPB marshals the ltfs.Directory structure to a pb.Entry. It does NOT
// recursively marshals the directory entries.
func MarshalDirectory(d *ltfs.Directory, entry *Entry) error {
	entry.Id = uint64(d.FileUID)
	entry.Name = d.Name

	entry.Readonly = d.ReadOnly
	entry.CreateTime = d.CreationTime.UnixNano()
	entry.ChangeTime = d.ChangeTime.UnixNano()
	entry.ModifyTime = d.ModifyTime.UnixNano()
	entry.AccessTime = d.AccessTime.UnixNano()
	entry.BackupTime = d.BackupTime.UnixNano()

	entry.Elem = &Entry_Dir{
		Dir: &Directory{},
	}

	return nil
}

func MarshalFile(f *ltfs.File, entry *Entry) error {
	entry.Id = uint64(f.FileUID)
	entry.Name = f.Name

	entry.Readonly = f.ReadOnly
	entry.CreateTime = f.CreationTime.UnixNano()
	entry.ChangeTime = f.ChangeTime.UnixNano()
	entry.ModifyTime = f.ModifyTime.UnixNano()
	entry.AccessTime = f.AccessTime.UnixNano()
	entry.BackupTime = f.BackupTime.UnixNano()

	pbfile := &File{
		Length:  uint64(f.Length),
		Extents: make([]*Extent, 0),
	}

	for _, extent := range f.ExtentInfo {
		pbfile.Extents = append(pbfile.Extents, &Extent{
			Id:        uint64(f.FileUID),
			Partition: 1, // TODO(kbj): don't hardcode this!
			Block:     uint64(extent.StartBlock),
			Length:    uint64(extent.ByteCount),
			Boffset:   uint64(extent.ByteOffset),
			Offset:    uint64(extent.FileOffset),
		})
	}

	entry.Elem = &Entry_File{pbfile}

	return nil
}
