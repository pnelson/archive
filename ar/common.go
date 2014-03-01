package ar

import (
	"os"
	"time"
)

const (
	magic     = "!<arch>\n" // The magic ASCII string required in the global header.
	fileMagic = "\x60\x0A"  // The magic file header terminator.
	blockSize = 60          // The size of the file header.
)

// A Header represents a single header in the ar archive.
type Header struct {
	Name    string    // The name of header file entry.
	ModTime time.Time // The modified time.
	Uid     int       // The user id of owner.
	Gid     int       // The group id of owner.
	Mode    int64     // The permission and mode bits.
	Size    int64     // The file size in bytes.
}

// FileInfo returns an os.FileInfo for the Header.
func (h *Header) FileInfo() os.FileInfo {
	return headerFileInfo{h}
}

// headerFileInfo implements os.FileInfo.
type headerFileInfo struct {
	h *Header
}

func (fi headerFileInfo) Name() string       { return fi.h.Name }
func (fi headerFileInfo) Size() int64        { return fi.h.Size }
func (fi headerFileInfo) Mode() os.FileMode  { return os.FileMode(fi.h.Mode).Perm() }
func (fi headerFileInfo) ModTime() time.Time { return fi.h.ModTime }
func (fi headerFileInfo) IsDir() bool        { return false }
func (fi headerFileInfo) Sys() interface{}   { return fi.h }
