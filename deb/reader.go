package deb

import (
	"archive/tar"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/pnelson/archive/ar"
)

var (
	ErrControl = errors.New("archive/deb: empty control file")
	ErrData    = errors.New("archive/deb: empty data file")
	ErrFormat  = errors.New("archive/deb: data format not implemented")
	ErrField   = errors.New("archive/deb: invalid control field")
)

// NewPackage creates a new Package from the file specified by path.
func NewPackage(path string) (*Package, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	pkg := &Package{
		Path:   path,
		Name:   fi.Name(),
		Size:   fi.Size(),
		Fields: make(map[string]string),
	}

	ar := ar.NewReader(f)
	for {
		hdr, err := ar.Next()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		switch hdr.Name {
		case "control.tar.gz":
			err = pkg.readControl(ar)
		case "data.tar":
			err = ErrFormat
		case "data.tar.bz2":
			err = ErrFormat
		case "data.tar.gz":
			r, err := NewTarGzReader(ar)
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				return nil, ErrData
			}

			err = pkg.readData(r)
		case "data.tar.lzma":
			err = ErrFormat
		case "data.tar.xz":
			err = ErrFormat
		}

		if err != nil {
			return nil, err
		}
	}

	err = pkg.parseFields()
	if err != nil {
		return nil, err
	}

	err = pkg.generateChecksums()
	if err != nil {
		return nil, err
	}

	sort.Strings(pkg.Files)

	return pkg, nil
}

// Reads raw control data from the package.
func (p *Package) readControl(ar *ar.Reader) error {
	r, err := NewTarGzReader(ar)
	if err != nil {
		if err != io.EOF {
			return err
		}
		return ErrControl
	}

	for {
		hdr, err := r.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		switch hdr.Name {
		case "./control":
			p.ControlData = string(buf)
		case "./preinst":
			p.PreInstData = string(buf)
		case "./prerm":
			p.PreRmData = string(buf)
		case "./postinst":
			p.PostInstData = string(buf)
		case "./postrm":
			p.PostRmData = string(buf)
		}
	}

	return nil
}

// Reads the data file for a listing of files the package installs.
func (p *Package) readData(r *tar.Reader) error {
	for {
		hdr, err := r.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		fi := hdr.FileInfo()
		if !fi.IsDir() {
			p.Files = append(p.Files, hdr.Name)
		}
	}

	return nil
}

// Parses the raw ControlData into individual fields.
func (p *Package) parseFields() error {
	prev := ""
	lines := strings.Split(p.ControlData, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if line[0] == ' ' || line[0] == '\t' {
			p.Fields[prev] += fmt.Sprintf("\n%s", line[1:])
			continue
		}

		parts := strings.SplitN(trimmed, ": ", 2)
		if len(parts) != 2 {
			return ErrField
		}

		key := parts[0]
		value := parts[1]
		p.Fields[key] = value
		prev = key
	}

	return nil
}

// Generates MD5, SHA1, and SHA256 checksums of the package.
func (p *Package) generateChecksums() error {
	buf, err := ioutil.ReadFile(p.Path)
	if err != nil {
		return err
	}

	p.Fields["MD5sum"] = generateChecksum(md5.New(), &buf)
	p.Fields["SHA1"] = generateChecksum(sha1.New(), &buf)
	p.Fields["SHA256"] = generateChecksum(sha256.New(), &buf)

	return nil
}

func generateChecksum(h hash.Hash, b *[]byte) string {
	_, err := h.Write(*b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
