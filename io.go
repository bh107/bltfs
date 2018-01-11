package bltfs

import (
	"bytes"
	"io"
)

// Copy copies data from the io.Reader to the io.Writer, ensuring that the
// write happens in full block sizes.
func (b *Store) Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, b.sopts.blkSize)

	var rerr, werr error

	for {
		var nr int

		// try to read at least blkSize bytes from the src
		for nr < int(b.sopts.blkSize) && rerr == nil {
			var nn int
			nn, rerr = src.Read(buf[nr:])
			nr += nn
		}

		if nr > 0 {
			// write as much as we read to dst
			var nw int
			nw, werr = dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}

			if werr != nil {
				err = werr
				break
			}

			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}

		if rerr != nil {
			if rerr != io.EOF {
				err = rerr
			}

			break
		}
	}

	return written, err
}

// WriteFile reads from the reader until EOF and writes to the underlying
// device. It does NOT write a filemark.
func (b *Store) WriteFile(r io.Reader) (int, error) {
	buf := make([]byte, b.sopts.blkSize)

	var written int
	for {
		n, err := io.ReadAtLeast(r, buf, int(b.sopts.blkSize))

		if err != nil {
			// we handle EOF and UnexpectedEOF in the same way
			if err == io.EOF && err == io.ErrUnexpectedEOF {
				break
			}

			return written, err
		}

		// write the block
		n, err = b.mu.backend.Write(buf)
		written += n
	}

	return written, nil
}

// ReadFile reads from the underlying device until the next filemark.
func (b *Store) ReadFile() ([]byte, error) {
	var buf bytes.Buffer
	var readZero bool

	// one device block
	blk := make([]byte, 524288)

	// read all records until next filemark
	for {
		n, err := b.mu.backend.Read(blk)

		if n == 0 {
			// nothing was read

			if err == nil {
				if readZero {
					// we read zero bytes last time, this means EOD
					return nil, ErrEOD
				}

				// no error, we have read the full file
				break
			}

			// we read zero bytes, make a note of it so we can catch the EOD signal
			readZero = true
		}

		if err != nil {
			return nil, err
		}

		// reset readZero
		readZero = false

		// write block to buffer
		buf.Write(blk)
	}

	return buf.Bytes(), nil
}
