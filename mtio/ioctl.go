package mtio

import (
	"os"
	"syscall"
)

func ioctl(fd, request, argp uintptr) error {
	_, _, errorp := syscall.Syscall(syscall.SYS_IOCTL, fd, request, argp)
	return os.NewSyscallError("ioctl", errorp)
}
