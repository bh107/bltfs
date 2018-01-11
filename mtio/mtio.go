package mtio

import (
	"unsafe"
)

const (
	MTSETPART = 33
)

func SetPartition(fd uintptr, part int32) error {
	op := MTOperation{
		Op:    MTSETPART,
		Count: part,
	}

	if err := ioctl(fd, uintptr(MTIOCTOP), uintptr(unsafe.Pointer(&op))); err != nil {
		return err
	}

	return nil
}
