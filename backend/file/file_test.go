package file

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/pkg/errors"
	"hpt.space/bltfs"
)

const (
	testDirectory = "/tmp/ltfs/tape"
)

func setup() string {
	dir, err := ioutil.TempDir("", "bltfstest")
	if err != nil {
		panic(err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	return dir
}

func cleanup(path string) {
	if err := os.RemoveAll(path); err != nil {
		panic(err)
	}
}

func TestReadWrite(t *testing.T) {
	dir := setup()

	dev, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := dev.Load(); err != nil {
		t.Fatal(err)
	}

	if err := dev.Format(); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 512*1024)

	sha256sums := make([][sha256.Size]byte, 4)

	for i := 0; i < len(sha256sums); i++ {
		n, err := rand.Read(buf)
		if err != nil || n != len(buf) {
			panic("should not (cannot!) happen")
		}

		sha256sums[i] = sha256.Sum256(buf)

		n, err = dev.Write(buf)
		if err != nil {
			t.Fatal(err)
		}

		if n != len(buf) {
			t.Fatal("n != len(buf)")
		}
	}

	if err := dev.WriteFilemark(1); err != nil {
		t.Fatal(err)
	}

	if err := dev.Close(); err != nil {
		t.Fatal(err)
	}

	dev, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := dev.Load(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(sha256sums); i++ {
		n, err := dev.Read(buf)
		if err != nil {
			t.Fatal(err)
		}

		if n != cap(buf) {
			t.Fatal("n != cap(buf)")
		}

		sum := sha256.Sum256(buf)

		if sum != sha256sums[i] {
			t.Fatal("sha256 mismatch")
		}
	}

	cleanup(dir)
}

func TestWriteFile(t *testing.T) {
	dir := setup()

	dev, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := dev.Load(); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open("/dev/urandom")
	if err != nil {
		t.Fatal(err)
	}

	n, err := io.CopyN(dev, f, 512*1024*4)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("wrote %d bytes\n", n)

	if err := dev.Close(); err != nil {
		t.Fatal(err)
	}

	cleanup(dir)
}

func TestOpen(t *testing.T) {
	dir := setup()

	_, err := Open("not_found")
	if err == nil {
		t.Fatal("expected an error")
	}

	dev, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	_ = dev

	cleanup(dir)
}

func TestLoad(t *testing.T) {
	dir := setup()

	dev, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := dev.Load(); err != nil {
		t.Fatal(err)
	}

	cleanup(dir)
}

func TestWriteFilemark(t *testing.T) {
	dir := setup()

	dev, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := dev.Load(); err != nil {
		t.Fatal(err)
	}

	if err := dev.WriteFilemark(1); err != nil {
		t.Fatal(err)
	}

	if err := dev.Close(); err != nil {
		t.Fatal(err)
	}

	cleanup(dir)
}

func TestReadFilemark(t *testing.T) {
	dir := setup()

	dev, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := dev.Load(); err != nil {
		t.Fatal(err)
	}

	if err := dev.WriteFilemark(1); err != nil {
		t.Fatal(err)
	}

	if err := dev.Close(); err != nil {
		t.Fatal(err)
	}

	dev, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := dev.Load(); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 512*1024)
	n, err := dev.Read(buf)
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to read from device"))
	}

	if n != 0 {
		t.Fatal("expected n = 0")
	}

	// next repeated read should return an error (end of device)
	for i := 0; i < 10; i++ {
		_, err = dev.Read(buf)
		if err != bltfs.ErrEOD {
			t.Fatal(err)
		}
	}

	if err := dev.Close(); err != nil {
		t.Fatal(err)
	}

	cleanup(dir)
}
