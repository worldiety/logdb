package logdb

import (
	"fmt"
	"github.com/worldiety/ioutil"
	"sync"
	"sync/atomic"
)

var headerMagic = [8]byte{'w', 'd', 'y', 'l', 'o', 'g', 'd', 'b'}

const headerVersion = 1

// Header marks the beginning of the database and provides space to organize names and indices.
type Header struct {
	buf             *ioutil.LittleEndianBuffer
	magic           [8]byte        // wdylogdb
	version         uint32         // currently only 1
	headerSize      uint32         // the total reserved size of the header. This determines the maximum amount of the string table size
	objCount        uint64         // the amount of objects
	txCount         uint64         // the amount of transactions
	nameCount       uint64         // amount of names
	lookup          map[string]int // reverse lookup from string to name index
	names           []string       // lookup index to string
	actualUsedBytes int
	mutex           sync.RWMutex
}

func newHeader(size int) *Header {
	h := &Header{
		buf: &ioutil.LittleEndianBuffer{
			Bytes: make([]byte, size),
			Pos:   0,
		},
		magic:           headerMagic,
		version:         headerVersion,
		objCount:        0,
		txCount:         0,
		nameCount:       0,
		headerSize:      0,
		lookup:          make(map[string]int),
		names:           nil,
		actualUsedBytes: 8 + 4 + 4 + 8 + 8 + 8,
	}
	return h
}

func (h *Header) ObjectCount() uint64 {
	return atomic.LoadUint64(&h.objCount)
}

func (h *Header) AddObjectCount(d uint64) uint64 {
	return atomic.AddUint64(&h.objCount, d)
}

func (h *Header) TxCount() uint64 {
	return atomic.LoadUint64(&h.txCount)
}

func (h *Header) AddTxCount(d uint64) uint64 {
	return atomic.AddUint64(&h.txCount, d)
}

func (h *Header) NameByIndex(idx int) string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.names[idx]
}

func (h *Header) IndexByName(name string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	idx, ok := h.lookup[name]
	if !ok {
		return -1
	}

	return idx
}

func (h *Header) Names() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	r := make([]string, len(h.names))
	for i, v := range h.names {
		r[i] = v
	}
	return r
}

func (h *Header) AddName(name string) int {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	idx, ok := h.lookup[name]
	if !ok {

		requiredSize := len(name) + 1 // + type
		if len(name) > int(ioutil.MaxUint8) {
			requiredSize += 1
		} else if len(name) > int(ioutil.MaxUint16) {
			requiredSize += 2
		} else if len(name) > int(ioutil.MaxUint24) {
			requiredSize += 3
		} else {
			requiredSize += 4
		}

		if h.actualUsedBytes+requiredSize > len(h.buf.Bytes) {
			panic(fmt.Sprintf("header overflow: has %d, needs another %d which exceeds %d", h.actualUsedBytes, requiredSize, len(h.buf.Bytes)))
		}

		h.actualUsedBytes += requiredSize

		h.names = append(h.names, name)
		idx = len(h.names) - 1
		h.lookup[name] = idx
		h.nameCount++

	}

	return idx
}

func (h *Header) reverseFlush() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.buf.Pos = 0
	h.buf.ReadSlice(h.magic[:])
	h.version = h.buf.ReadUint32()
	h.headerSize = h.buf.ReadUint32()
	h.objCount = h.buf.ReadUint64()
	h.txCount = h.buf.ReadUint64()
	h.nameCount = h.buf.ReadUint64()

	// this is actually an optimized clear-map, see https://github.com/golang/go/issues/20138
	for k := range h.lookup {
		delete(h.lookup, k)
	}

	if len(h.names) != int(h.nameCount) {
		h.names = make([]string, h.nameCount)
	}

	for i := range h.names {
		h.names[i] = ""
	}

	for i := 0; i < int(h.nameCount); i++ {
		name := (*ioutil.TypedLittleEndianBuffer)(h.buf).ReadString(nil)
		h.names[i] = name
		h.lookup[name] = i
	}

	h.actualUsedBytes = h.buf.Pos
}

func (h *Header) Flush() {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	h.buf.Pos = 0
	h.buf.WriteSlice(h.magic[:])
	h.buf.WriteUint32(h.version)
	h.buf.WriteUint32(h.headerSize)
	h.buf.WriteUint64(h.objCount)
	h.buf.WriteUint64(h.txCount)
	h.buf.WriteUint64(h.nameCount)

	for _, name := range h.names {
		(*ioutil.TypedLittleEndianBuffer)(h.buf).WriteString(name)
	}

	h.actualUsedBytes = h.buf.Pos
}
