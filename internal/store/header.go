package store

import (
	"fmt"
	"github.com/worldiety/ioutil"
)

// magic contains the header magic bytes
var magic = [8]byte{'w', 'd', 'y', 'l', 'o', 'g', 'd', 'b'}

// current binary version
const version = 1

const maxHeaderSize = 8 + 4 + 8 + 8 + 2 + ioutil.MaxUint16

// Header of the store, which is at most 16.777.238 Byte.
type Header struct {
	magic           [8]byte  //wdylogdb
	version         uint32   // 1
	entries         uint64   // row count
	lastTransaction uint64   // strict monotonic increasing number
	fieldCount      uint16   // for the names
	fieldNames      []string // each at most 255 chars
}

// NewHeader creates an empty header
func NewHeader() Header {
	return Header{
		magic:           magic,
		version:         version,
		entries:         0,
		lastTransaction: 0,
		fieldCount:      0,
	}
}

// ParseHeader reads from the current position.
func ParseHeader(r ioutil.DataInput) (Header, error) {
	h := Header{}
	r.ReadFull(h.magic[:])
	h.version = r.ReadUint32()
	h.entries = r.ReadUint64()
	h.lastTransaction = r.ReadUint64()
	h.fieldCount = r.ReadUint16()
	h.fieldNames = make([]string, h.fieldCount)
	for i := uint16(0); i < h.fieldCount; i++ {
		h.fieldNames[i] = r.ReadUTF8(ioutil.I8)
	}

	if h.magic != magic {
		return h, fmt.Errorf("unexpected header magic: %v, expected %v", h.magic, magic)
	}

	if h.version != version {
		return h, fmt.Errorf("invalid header version: %d expected %d", h.version, version)
	}

	return h, r.Error()
}

// Write emits the header at the current position.
func (h Header) Write(w ioutil.DataOutput) error {
	_, _ = w.Write(h.magic[:])
	w.WriteUint32(h.version)
	w.WriteUint64(h.entries)
	w.WriteUint64(h.lastTransaction)
	w.WriteUint16(h.fieldCount)
	for _, name := range h.fieldNames {
		w.WriteUTF8(ioutil.I8, name)
	}

	return w.Error()
}

// Field returns -1 or the index of the name, if any. This is a o(n) linear lookup, so be sure to cache
// the results or do it only once before scanning.
func (h Header) Field(name string) int {
	for i, v := range h.fieldNames {
		if v == name {
			return i
		}
	}

	return -1
}

// AddField inserts another field. If field already exists, this is a no-op.
func (h *Header) AddField(name string) int {
	for i, v := range h.fieldNames {
		if v == name {
			return i
		}
	}

	h.fieldCount++
	h.fieldNames = append(h.fieldNames, name)
	return len(h.fieldNames) - 1
}

// Equals is only true, if the headers are logically equivalent
func (h Header) Equals(other Header) bool {
	if h.magic != other.magic {
		return false
	}

	if h.version != other.version {
		return false
	}

	if h.entries != other.entries {
		return false
	}

	if h.lastTransaction != other.lastTransaction {
		return false
	}

	if h.fieldCount != other.fieldCount {
		return false
	}

	if len(h.fieldNames) != len(h.fieldNames) {
		return false
	}

	for i := 0; i < len(h.fieldNames); i++ {
		if h.fieldNames[i] != other.fieldNames[i] {
			return false
		}
	}

	return true

}
