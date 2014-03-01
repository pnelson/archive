package ar

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

// A Reader provides sequential access to the contents of an ar archive.
// An ar archive consists of a sequence of files.
// The Next method advances to the next file in the archive (including the
// first), and then it can be treated as an io.Reader to access the file's
// data.
type Reader struct {
	r      io.Reader
	err    error
	valid  bool
	padded bool
	data   io.LimitedReader
}

var (
	ErrMagic  = errors.New("archive/ar: invalid ar global header magic string")
	ErrHeader = errors.New("archive/ar: invalid ar header")
)

// NewReader creates a new Reader reading from r.
func NewReader(r io.Reader) *Reader {
	return &Reader{r: r}
}

// Next advances to the next entry in the ar archive, including the first.
func (ar *Reader) Next() (*Header, error) {
	if ar.err == nil {
		ar.skipUnread()
	}

	if ar.err != nil {
		return nil, ar.err
	}

	if !ar.valid {
		if err := ar.validate(); err != nil {
			return nil, ErrMagic
		}
	}

	header := ar.readHeader()
	if header == nil {
		return nil, ar.err
	}

	return header, nil
}

// Read reads from the current entry in the ar archive.
func (ar *Reader) Read(b []byte) (int, error) {
	if ar.err != nil {
		return 0, ar.err
	}

	if ar.data.R == nil {
		return 0, os.ErrNotExist
	}

	return ar.data.Read(b)
}

// Parses a single ar file header into a Header.
func (ar *Reader) readHeader() *Header {
	raw := make([]byte, blockSize)
	if _, err := io.ReadFull(ar.r, raw); err != nil {
		ar.err = err
		return nil
	}

	if string(raw[blockSize-2:blockSize]) != fileMagic {
		ar.err = ErrHeader
		return nil
	}

	header := new(Header)
	header.Name = string(bytes.TrimSpace(raw[0:16]))
	header.ModTime = ar.modTime(raw[16:28])
	header.Uid = int(ar.readInt(raw[28:34], 10))
	header.Gid = int(ar.readInt(raw[34:40], 10))
	header.Mode = ar.readInt(raw[40:48], 8)
	header.Size = ar.readInt(raw[48:58], 10)

	if ar.err != nil {
		ar.err = ErrHeader
		return nil
	}

	ar.data.R = ar.r
	ar.data.N = header.Size

	ar.padded = header.Size%2 == 1

	return header
}

// Reads an integer in a given base from a slice of bytes, ignoring padding.
func (ar *Reader) readInt(b []byte, base int) int64 {
	s := string(bytes.TrimSpace(b))
	rv, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		ar.err = err
		return 0
	}
	return rv
}

// Reads a unix timestamp from a slice of bytes.
func (ar *Reader) modTime(b []byte) time.Time {
	rv := ar.readInt(b, 10)
	if rv == 0 {
		return time.Time{}
	}
	return time.Unix(rv, 0)
}

// Skips unread bytes in the existing file entry.
func (ar *Reader) skipUnread() {
	_, ar.err = io.Copy(ioutil.Discard, &ar.data)
	if ar.padded {
		c := make([]byte, 1)
		_, ar.err = ar.r.Read(c)
		if c[0] != '\n' {
			ar.err = ErrHeader
		}
	}
}

// Validates the magic ASCII string that the global header expects.
func (ar *Reader) validate() error {
	_, err := io.CopyN(ioutil.Discard, ar.r, int64(len(magic)))
	if err != nil {
		return err
	}
	ar.valid = true
	return nil
}
