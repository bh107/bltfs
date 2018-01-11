// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs ctypes_mtio.go

package mtio

type MTOperation struct {
	Op		int16
	Pad_cgo_0	[2]byte
	Count		int32
}
type MTGet struct {
	Type	int64
	Resid	int64
	Dsreg	int64
	Gstat	int64
	Erreg	int64
	Fileno	int32
	Blkno	int32
}
type MTPos struct {
	Blkno int64
}

const Sizeof_MTOperation = 0x8
const Sizeof_MTGet = 0x30
const Sizeof_MTPos = 0x8
