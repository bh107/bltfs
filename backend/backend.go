package backend

// Interface is the interface that bLTFS backends must implement.
type Interface interface {
	BlockSize() uint64

	// While discouraged by the io.Reader interface, note that Read may return a
	// zero byte count with a nil error. The following is quoted verbatim from
	// the Linux st(4) man page.
	//
	// When a filemark is encountered while reading, the following happens. If
	// there are data remaining in the buffer when the filemark is found, the
	// buffered data is returned. The next read returns zero bytes. The following
	// read returns data from the next file. The end of recorded data is signaled
	// by returning zero bytes for two consecutive read calls. The third read
	// returns an error.
	Read(p []byte) (int, error)

	// Write writes up to the maximum block size of the underlying device. If the
	// device block size is less than len(p) it returns the number of bytes
	// written and the error ErrBlockSizeExceeded.
	Write(p []byte) (int, error)

	// Write count number of file marks to the device.
	WriteFilemark(count int) error

	// Format the device (currently a no-op).
	Format() error

	Close() error

	// Rewind the device.
	Rewind() error

	// Load initializes the device.
	Load() error

	// Unload unloads the device (currently a no-op).
	Unload() error

	// Seek to the specified position.
	Locate(part uint32, block uint64) error

	// Space over the device to EOD.
	SpaceEOD() error

	// Forward space count files. The tape is positioned on the first block of
	// the next file.
	SpaceFMF(count uint64) error

	// Backward space count files. The tape is positioned on the first block of
	// the next file.
	SpaceFMB(count uint64) error

	// ReadPosition returns the current position of the drive as a logical block
	// number.
	ReadPosition() (uint64, error)

	// SetPartitions sets the active partition.
	SetPartition(part uint32) error
}
