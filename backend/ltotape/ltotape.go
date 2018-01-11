package ltotape

/*
import (
	"os"

	"github.com/tapebit/bltfs/backend"
	"github.com/tapebit/bltfs/mtio"
)

// ensure that the Device type implements backend.Interface
var _ backend.Interface = &Device{}

type Device struct {
	f *os.File
}

func Open(tapedev string) (*Device, error) {
	f, err := os.OpenFile(tapedev, os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}

	return &Device{
		f: f,
	}, nil
}

func (d *Device) Close() error {
	return d.f.Close()
}

func (d *Device) Read(p []byte) (n int, err error) {
	return d.f.Read(p)
}

func (d *Device) Write(p []byte) (n int, err error) {
	return d.f.Write(p)
}

func (d *Device) WriteFilemark(count int) error {
	return nil
}

func (d *Device) Rewind() error {
	return nil
}

func (d *Device) Format() error {
	return nil
}

func (d *Device) Load() error {
	return nil
}

func (d *Device) Unload() error {
	return nil
}

func (d *Device) SetPartition(part int32) error {
	return mtio.SetPartition(d.f.Fd(), part)
}
*/
