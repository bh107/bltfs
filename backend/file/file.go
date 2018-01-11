// Package file implements a file-based tape emulator.
//
// Basically ported from tape_drivers/generic/file/filedebug_tc.c from the
// IBM LTFS SDE distribution, this emulator uses the same format as IBM which
// allows the IBM LTFS utilities to work with emulated tape volumes created
// with this implementation.
package file

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"hpt.space/bltfs"
	"hpt.space/bltfs/backend"
	"hpt.space/bltfs/util/fsutil"
)

const (
	SuffixRecord   = "R"
	SuffixFilemark = "F"
	SuffixEOD      = "E"
	SuffixANY      = "."

	EODMissing = math.MaxUint64
)
const (
	DefaultCapacity  = 3 * 1024
	DefaultBlockSize = 512 * 1024
)

var _ backend.Interface = &device{}

type position struct {
	blk  uint64
	part uint32
}

type capacity struct {
	p0 struct {
		remaining uint64
		max       uint64
	}

	p1 struct {
		remaining uint64
		max       uint64
	}
}

type device struct {
	root       string
	blkSize    uint64
	pos        *position
	partitions uint32

	last []uint64
	eod  []uint64

	ready bool

	cartCfg *CartridgeConfig
}

func (p *position) adv(count uint64) {
	p.blk += count
}

func (p *position) rev(count uint64) {
	if p.blk-count < 0 {
		panic("reversing to negative block")
	}

	p.blk -= count
}

func (p *position) reset() {
	p.blk = 0
	p.part = 0
}

func Open(root string) (backend.Interface, error) {
	finfo, err := os.Stat(root)
	if err != nil {
		return nil, err
	}

	if !finfo.IsDir() {
		return nil, errors.New("path must be an existing directory")
	}

	return &device{
		cartCfg: &CartridgeConfig{
			DummyIO:         false,
			EmulateReadOnly: false,
			Capacity:        DefaultCapacity,
			CartridgeType:   "L5",
			DensityCode:     0x58,
		},

		root:    root,
		blkSize: DefaultBlockSize,
		pos:     &position{},

		partitions: 2,
		last:       make([]uint64, 2),
		eod:        make([]uint64, 2),
	}, nil
}

func (d *device) BlockSize() uint64 {
	return d.blkSize
}

func (d *device) Close() error {
	return nil
}

func (d *device) Load() error {
	if d.ready {
		d.pos.reset()

		return nil
	}

	cartCfg, err := readCartridgeConfig(filepath.Join(d.root, DefaultCartridgeConfigFile))
	if err != nil {
		if os.IsNotExist(err) {
			if err := writeCartridgeConfig(filepath.Join(d.root, DefaultCartridgeConfigFile), d.cartCfg); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	d.cartCfg = cartCfg

	d.ready = true

	for part := range d.eod {
		if err := d.searchEOD(uint32(part)); err != nil {
			return err
		}
	}

	// rewind-ish
	d.pos.reset()

	return nil
}

func (d *device) Unload() error {
	d.ready = false

	d.pos.reset()

	return nil
}

func (d *device) SetPartition(part uint32) error {
	d.pos.part = part
	return nil
}

func (d *device) Format() error {
	if d.pos.part != 0 || d.pos.blk != 0 {
		return errors.New("illegal request")
	}

	return nil
}

func (d *device) Rewind() error {
	d.pos.blk = 0

	return nil
}

func (d *device) ReadPosition() (uint64, error) {
	return d.pos.blk, nil
}

func (d *device) onFilemark() bool {
	path := d.makeFilemarkPath(d.pos)
	return fsutil.Exists(path)
}

func (d *device) onRecord() bool {
	path := d.makeRecordPath(d.pos)
	return fsutil.Exists(path)
}

func (d *device) Read(p []byte) (int, error) {
	if !d.ready {
		return 0, bltfs.ErrNotReady
	}

	if len(p) < int(d.blkSize) {
		return 0, io.ErrShortBuffer
	}

	if d.eod[d.pos.part] == d.pos.blk {
		return 0, bltfs.ErrEOD
	}

	// check for filemark (returns 0 bytes and advanced position)
	if d.onFilemark() {
		d.pos.adv(1)
		return 0, nil
	}

	// check that we are on a record
	if !d.onRecord() {
		return 0, errors.New("no such record")
	}

	f, err := os.Open(d.makeRecordPath(d.pos))
	if err != nil {
		return 0, err
	}

	defer f.Close()
	defer d.pos.adv(1)

	buf := p
	var total int
	for {
		n, err := f.Read(buf)

		buf = buf[:n]

		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}

			return total, err
		}

		total += n
	}

	return total, nil
}

func (d *device) Write(p []byte) (n int, err error) {
	if !d.ready {
		return 0, bltfs.ErrNotReady
	}

	if err := d.cleanCurrent(); err != nil {
		return 0, err
	}

	path := d.makeRecordPath(d.pos)

	buf := p

	// write at most up to the device block size
	if len(p) > int(d.blkSize) {
		buf = p[:d.blkSize]
		err = io.ErrShortWrite
	}

	if err := ioutil.WriteFile(path, buf, 0644); err != nil {
		return 0, err
	}

	// advance tape position
	d.pos.adv(1)

	if err := d.writeEOD(); err != nil {
		return len(buf), err
	}

	n = len(buf)

	return
}

func (d *device) WriteFilemark(count int) error {
	if !d.ready {
		return bltfs.ErrNotReady
	}

	for i := 0; i < count; i++ {
		// clean-up anything previously in this block
		if err := d.cleanCurrent(); err != nil {
			return errors.Wrapf(err, "failed to remove record at current position (%v)", d.pos)
		}

		path := d.makeFilemarkPath(d.pos)

		f, err := os.Create(path)
		if err != nil {
			return errors.Wrapf(err, "failed to create file %s", path)
		}

		f.Close()

		// advance to the next logical block
		d.pos.adv(1)

		if err := d.writeEOD(); err != nil {
			return err
		}
	}

	return nil
}

func (d *device) SpaceEOD() error {
	return nil
}

func (d *device) SpaceFMB(count uint64) error {
	if count == 0 {
		return nil
	}

	var n uint64
	for {
		path := d.makeFilemarkPath(d.pos)

		if fsutil.Exists(path) {
			n++
			if n == count {
				// advance to the first block of the next file
				d.pos.adv(1)
				return nil
			}
		}

		if d.pos.blk == 0 {
			return bltfs.ErrBOT
		}

		d.pos.rev(1)
	}
}

func (d *device) SpaceFMF(count uint64) error {
	if count == 0 {
		return nil
	}

	// check if current position is at EOD
	if d.pos.blk == d.eod[d.pos.part] {
		return bltfs.ErrEOD
	}

	// check if current position is last block before EOD
	if d.pos.blk == d.last[d.pos.part] {
		return bltfs.ErrIO
	}

	// TODO(kbj): we don't wanna risk an endless loop here.
	if d.pos.blk > d.last[d.pos.part] {
		panic("eh, panic")
	}

	var n uint64
	for {
		path := d.makeFilemarkPath(d.pos)

		if fsutil.Exists(path) {
			n++
			if n == count {
				d.pos.adv(1)
				return nil
			}
		}

		d.pos.adv(1)
	}
}

func (d *device) Locate(part uint32, block uint64) error {
	if !d.ready {
		return bltfs.ErrNotReady
	}

	d.pos.part = part
	if d.eod[part] == EODMissing && d.last[part] < block {
		d.pos.blk = d.last[part] + 1
	} else if d.eod[part] < block {
		d.pos.blk = d.eod[part]
	} else {
		d.pos.blk = block
	}

	return nil
}

func (d *device) capacity(part int) uint64 {
	switch part {
	case 0:
		return d.cartCfg.Capacity * 5 / 100
	case 1:
		return d.cartCfg.Capacity - d.capacity(0)
	default:
		panic("no more that two partitions supported")
	}
}

func (d *device) writeEOD() error {
	if err := d.cleanCurrent(); err != nil {
		return err
	}

	path := d.makeEODPath(d.pos)

	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", path)
	}

	f.Close()

	// remove any records following this position
	for i := d.pos.blk + 1; i <= d.eod[d.pos.part]; i++ {
		if err := d.clean(&position{blk: i, part: d.pos.part}); err != nil {
			return err
		}
	}

	d.last[d.pos.part] = d.pos.blk - 1
	d.eod[d.pos.part] = d.pos.blk

	return nil
}

func (d *device) clean(p *position) error {
	path := d.makePath(p, SuffixANY)
	for _, suffix := range []string{SuffixRecord, SuffixFilemark, SuffixEOD} {
		path := path[:len(path)-1]

		// ignore IsNotExists error
		if err := os.Remove(path + suffix); err != nil && !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to clean position (%d:%d)", p.part, p.blk)
		}
	}

	return nil
}

func (d *device) cleanCurrent() error {
	return d.clean(d.pos)
}

func (d *device) makePath(p *position, suffix string) string {
	s := fmt.Sprintf("%d_%d_%s", p.part, p.blk, suffix)
	return filepath.Join(d.root, s)
}

func (d *device) makeRecordPath(p *position) string {
	return d.makePath(p, SuffixRecord)
}

func (d *device) makeFilemarkPath(p *position) string {
	return d.makePath(p, SuffixFilemark)
}

func (d *device) makeEODPath(p *position) string {
	return d.makePath(p, SuffixEOD)
}

// The algorithm for this is ported from the autistic version from
// tape_drivers/generic/file/filedebug_tc.c in the IBM LTFS distribution.
//
// Mad probs to the authors Atsushi Abe & Brian Biskeborn.
func (d *device) searchEOD(part uint32) error {
	d.pos.reset()
	d.pos.part = part

	found := map[string]bool{
		SuffixRecord:   true,
		SuffixFilemark: true,
		SuffixEOD:      false,
	}

	for (found[SuffixRecord] || found[SuffixFilemark]) && !found[SuffixEOD] {
		path := d.makePath(d.pos, SuffixANY)

		for _, suffix := range []string{SuffixRecord, SuffixFilemark, SuffixEOD} {
			path := path[:len(path)-1]

			found[suffix] = fsutil.Exists(path + suffix)
		}

		d.pos.adv(1)
	}

	d.pos.rev(1)

	if !found[SuffixEOD] && d.pos.blk != 0 {
		d.last[part] = d.pos.blk
		d.eod[part] = EODMissing
	} else {
		if err := d.writeEOD(); err != nil {
			return err
		}
	}

	return nil
}
