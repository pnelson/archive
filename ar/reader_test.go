package ar

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

// Ensure that headerFileInfo implements os.FileInfo.
var _ os.FileInfo = new(headerFileInfo)

var arTest = []struct {
	file    string
	headers []*Header
	data    []string
}{
	{
		file: "testdata/empty.deb",
		headers: []*Header{
			{
				Name:    "debian-binary",
				ModTime: time.Unix(1393442320, 0),
				Uid:     0,
				Gid:     0,
				Mode:    0100644,
				Size:    4,
			},
			{
				Name:    "control.tar.gz",
				ModTime: time.Unix(1393442320, 0),
				Uid:     0,
				Gid:     0,
				Mode:    0100644,
				Size:    0,
			},
			{
				Name:    "data.tar.lzma",
				ModTime: time.Unix(1393442320, 0),
				Uid:     0,
				Gid:     0,
				Mode:    0100644,
				Size:    0,
			},
		},
		data: []string{
			"2.0\n",
			"",
			"",
		},
	},
}

func TestReader(t *testing.T) {
testLoop:
	for i, test := range arTest {
		f, err := os.Open(test.file)
		if err != nil {
			t.Errorf("%d. unexpected error: %v", i, err)
			continue
		}
		defer f.Close()

		ar := NewReader(f)
		for j, header := range test.headers {
			hdr, err := ar.Next()
			if err != nil || hdr == nil {
				t.Errorf("%d. header %d not found: %v", i, j, err)
				f.Close()
				continue testLoop
			}
			if *hdr != *header {
				t.Errorf("%d. header %d incorrect;\nhave %+v\nwant %+v", i, j, *hdr, *header)
				continue
			}

			data, err := ioutil.ReadAll(ar)
			if err != nil {
				t.Errorf("%d. header %d data read error: %v", i, j, err)
				continue
			}

			want := test.data[j]
			if have := string(data); have != want {
				t.Errorf("%d. data %d incorrect; have %q, want %q", i, j, have, want)
			}
		}

		hdr, err := ar.Next()
		if err == io.EOF {
			continue testLoop
		}
		if hdr != nil || err != nil {
			t.Errorf("%d. unexpected header or error;\nhdr=%v\nerr=%v", i, hdr, err)
		}
	}
}

func TestFileInfo(t *testing.T) {
	f, err := os.Open(arTest[0].file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	ar := NewReader(f)
	hdr, err := ar.Next()
	if err != nil || hdr == nil {
		t.Fatalf("didn't get first file: %v", err)
	}

	fi := hdr.FileInfo()
	if name := fi.Name(); name != hdr.Name {
		t.Errorf("fi.Name() = %v; want %v", name, hdr.Name)
	}
	if size := fi.Size(); size != hdr.Size {
		t.Errorf("fi.Size() = %v; want %v", size, hdr.Size)
	}
	if mode := fi.Mode(); mode != os.FileMode(hdr.Mode).Perm() {
		t.Errorf("fi.Mode() = %v; want %v", mode, os.FileMode(hdr.Mode).Perm())
	}
	if modTime := fi.ModTime(); modTime != hdr.ModTime {
		t.Errorf("fi.ModTime() = %v; want %v", modTime, hdr.ModTime)
	}
	if isDir := fi.IsDir(); isDir != false {
		t.Errorf("fi.IsDir() = %v; want %v", isDir, false)
	}
}

func TestPartialRead(t *testing.T) {
	f, err := os.Open(arTest[0].file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	ar := NewReader(f)
	hdr, err := ar.Next()
	if err != nil || hdr == nil {
		t.Fatalf("didn't get first file: %v", err)
	}

	buf := make([]byte, 3)
	if _, err := io.ReadFull(ar, buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expected := []byte("2.0"); !bytes.Equal(buf, expected) {
		t.Errorf("contents = %v, want %v", buf, expected)
	}

	hdr, err = ar.Next()
	if err != nil || hdr == nil {
		t.Fatalf("didn't get second file: %v", err)
	}
}
