package store

import (
	"fmt"
	"github.com/worldiety/ioutil"
)

// Footer is written at the end of the store, so that the store knows that everything has been written properly.
type Footer struct {
	magic         [8]byte // wdylogdb
	version       uint32  // 1
	transactionId uint64  // this must be equal to the header, otherwise something went wrong
}

// NewFooter creates a new footer, ready to write out
func NewFooter(tx uint64) Footer {
	return Footer{
		magic:         magic,
		version:       version,
		transactionId: tx,
	}
}

// ParseFooter reads from the current position.
func ParseFooter(r ioutil.DataInput) (Footer, error) {
	h := Footer{}
	r.ReadFull(h.magic[:])
	h.version = r.ReadUint32()
	h.transactionId = r.ReadUint64()

	if h.magic != magic {
		return h, fmt.Errorf("unexpected footer magic: %v, expected %v", h.magic, magic)
	}

	if h.version != version {
		return h, fmt.Errorf("invalid footer version: %d expected %d", h.version, version)
	}

	return h, r.Error()
}

// Write emits the header at the current position.
func (h Footer) Write(w ioutil.DataOutput) error {
	_, _ = w.Write(h.magic[:])
	w.WriteUint32(h.version)
	w.WriteUint64(h.transactionId)
	return w.Error()
}

// TransactionId returns the footers id.
func (h Footer) TransactionId() uint64 {
	return h.transactionId
}
