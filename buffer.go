package logdb

import (
	"math"
	"reflect"
	"strconv"
	"unsafe"
)

const (
	// MaxUint8 is 255
	MaxUint8 = 1<<8 - 1

	// MaxUint16 is 65535
	MaxUint16 = 1<<16 - 1
	// MaxUint24 is 16777215
	MaxUint24 = 1<<24 - 1
	// MaxUint32 is 4294967295
	MaxUint32 = 1<<32 - 1
	// MaxUint40 is 1099511627775
	MaxUint40 = 1<<40 - 1
	// MaxUint48 is 281474976710655
	MaxUint48 = 1<<48 - 1
	// MaxUint56 is 72057594037927935
	MaxUint56 = 1<<56 - 1
	// MaxUint64 is 18446744073709551615
	MaxUint64 = 1<<64 - 1
)

type DataType byte

const (
	TUint8      DataType = 1
	TUint16     DataType = 2
	TUint24     DataType = 3
	TUint32     DataType = 4
	TUint64     DataType = 5
	TTinyBlob   DataType = 6
	TBlob       DataType = 7
	TMediumBlob DataType = 8
	TLongBlob   DataType = 9
	TFloat32    DataType = 10
	TFloat64    DataType = 11

	minTValid = TUint8
	maxTValid = TFloat64
)



// LEBuffer is a light weight helper to modify bytes within a buffer
type LEBuffer struct {
	Buf []byte
	Pos int
}

func (f *LEBuffer) ReadUint8() uint8 {
	b := f.Buf[f.Pos]
	f.Pos++
	return b
}

func (f *LEBuffer) WriteUint8(v uint8) {
	f.Buf[f.Pos] = v
	f.Pos++
}

func (f *LEBuffer) ReadUint16() uint16 {
	b := f.Buf[f.Pos:]
	f.Pos += 2
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[0]) | uint16(b[1])<<8
}

func (f *LEBuffer) WriteUint16(v uint16) {
	b := f.Buf[f.Pos:]
	f.Pos += 2
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

func (f *LEBuffer) ReadUint24() uint32 {
	b := f.Buf[f.Pos:]
	f.Pos += 3

	_ = b[2] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
}

func (f *LEBuffer) WriteUint24(v uint32) {
	b := f.Buf[f.Pos:]
	f.Pos += 3

	_ = b[2] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)  //nolint:gomnd
	b[2] = byte(v >> 16) //nolint:gomnd
}

func (f *LEBuffer) ReadUint32() uint32 {
	b := f.Buf[f.Pos:]
	f.Pos += 4
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func (f *LEBuffer) WriteUint32(v uint32) {
	b := f.Buf[f.Pos:]
	f.Pos += 4

	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func (f *LEBuffer) ReadUint64() uint64 {
	b := f.Buf[f.Pos:]
	f.Pos += 8

	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

func (f *LEBuffer) WriteUint64(v uint64) {
	b := f.Buf[f.Pos:]
	f.Pos += 8

	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

// WriteSlice copies the content of the given buffer into the destination
func (f *LEBuffer) WriteSlice(v []byte) {
	b := f.Buf[f.Pos : f.Pos+len(v)]
	copy(b, v)
	f.Pos += len(v)
}

// ReadSlice reads fully into the given buffer
func (f *LEBuffer) ReadSlice(v []byte) {
	b := f.Buf[f.Pos : f.Pos+len(v)]
	copy(v, b)
	f.Pos += len(v)
}

// ReadTinyBlob reads up to 255 bytes. The blob is truncated.
func (f *LEBuffer) ReadTinyBlob(v []byte) int {
	vLen := f.ReadUint8()
	vBuf := v[0:vLen]

	f.ReadSlice(vBuf)
	return int(vLen)
}

// WriteTinyBlob writes up to 255 bytes. The blob is truncated.
func (f *LEBuffer) WriteTinyBlob(v []byte) {
	vLen := len(v)
	if vLen > MaxUint8 {
		vLen = MaxUint8
	}

	f.WriteUint8(uint8(vLen))
	f.WriteSlice(v[:vLen])
}

// ReadBlob reads up to 65535 bytes. The blob is truncated.
func (f *LEBuffer) ReadBlob(v []byte) int {
	vLen := f.ReadUint16()
	vBuf := v[0:vLen]

	f.ReadSlice(vBuf)
	return int(vLen)
}

// WriteBlob writes up to 65535 bytes. The blob is truncated.
func (f *LEBuffer) WriteBlob(v []byte) {
	vLen := len(v)
	if vLen > MaxUint16 {
		vLen = MaxUint16
	}

	f.WriteUint16(uint16(vLen))
	f.WriteSlice(v[:vLen])
}

// ReadBlob reads up to 16777215 bytes. The blob is truncated.
func (f *LEBuffer) ReadMediumBlob(v []byte) int {
	vLen := f.ReadUint24()
	vBuf := v[0:vLen]

	f.ReadSlice(vBuf)
	return int(vLen)
}

// WriteBlob writes up to 16777215 bytes. The blob is truncated.
func (f *LEBuffer) WriteMediumBlob(v []byte) {
	vLen := len(v)
	if vLen > MaxUint24 {
		vLen = MaxUint24
	}

	f.WriteUint24(uint32(vLen))
	f.WriteSlice(v[:vLen])
}

// ReadLongBlob reads up to 4294967295 bytes. The blob is truncated.
func (f *LEBuffer) ReadLongBlob(v []byte) int {
	vLen := f.ReadUint32()
	vBuf := v[0:vLen]

	f.ReadSlice(vBuf)
	return int(vLen)
}

// WriteLongBlob writes up to 4294967295 bytes. The blob is truncated.
func (f *LEBuffer) WriteLongBlob(v []byte) {
	vLen := len(v)
	if vLen > MaxUint32 {
		vLen = MaxUint32
	}

	f.WriteUint32(uint32(vLen))
	f.WriteSlice(v[:vLen])
}

// WriteTinyString writes the string into a blob, avoiding another allocation.
func (f *LEBuffer) WriteTinyString(v string) {
	str := *(*reflect.StringHeader)(unsafe.Pointer(&v))
	// do not modify the slice, because this is a hack to avoid an unnecessary copy and heap allocation
	slice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: str.Data,
		Len:  str.Len,
		Cap:  str.Len,
	}))

	f.WriteTinyBlob(slice)
}

// ReadTinyString creates a (mutable) string, by using the strBuffer.
func (f *LEBuffer) ReadTinyString(strBuffer []byte) string {
	vLen := f.ReadTinyBlob(strBuffer)
	strBuffer = strBuffer[:vLen]
	// this hack avoids another allocation for the string, see https://github.com/golang/go/issues/25484
	return *(*string)(unsafe.Pointer(&vLen))
}

// WriteString writes the string into a blob, avoiding another allocation.
func (f *LEBuffer) WriteString(v string) {
	str := *(*reflect.StringHeader)(unsafe.Pointer(&v))
	// do not modify the slice, because this is a hack to avoid an unnecessary copy and heap allocation
	slice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: str.Data,
		Len:  str.Len,
		Cap:  str.Len,
	}))

	f.WriteBlob(slice)
}

// ReadString creates a (mutable) string, by using the strBuffer.
func (f *LEBuffer) ReadString(strBuffer []byte) string {
	vLen := f.ReadBlob(strBuffer)
	strBuffer = strBuffer[:vLen]
	// this hack avoids another allocation for the string, see https://github.com/golang/go/issues/25484
	return *(*string)(unsafe.Pointer(&vLen))
}

// WriteMediumString writes the string into a blob, avoiding another allocation.
func (f *LEBuffer) WriteMediumString(v string) {
	str := *(*reflect.StringHeader)(unsafe.Pointer(&v))
	// do not modify the slice, because this is a hack to avoid an unnecessary copy and heap allocation
	slice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: str.Data,
		Len:  str.Len,
		Cap:  str.Len,
	}))

	f.WriteMediumBlob(slice)
}

// ReadMediumString creates a (mutable) string, by using the strBuffer.
func (f *LEBuffer) ReadMediumString(strBuffer []byte) string {
	vLen := f.ReadMediumBlob(strBuffer)
	strBuffer = strBuffer[:vLen]
	// this hack avoids another allocation for the string, see https://github.com/golang/go/issues/25484
	return *(*string)(unsafe.Pointer(&vLen))
}

// WriteLongString writes the string into a blob, avoiding another allocation.
func (f *LEBuffer) WriteLongString(v string) {
	str := *(*reflect.StringHeader)(unsafe.Pointer(&v))
	// do not modify the slice, because this is a hack to avoid an unnecessary copy and heap allocation
	slice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: str.Data,
		Len:  str.Len,
		Cap:  str.Len,
	}))

	f.WriteLongBlob(slice)
}

// ReadLongString creates a (mutable) string, by using the strBuffer.
func (f *LEBuffer) ReadLongString(strBuffer []byte) string {
	vLen := f.ReadLongBlob(strBuffer)
	strBuffer = strBuffer[:vLen]
	// this hack avoids another allocation for the string, see https://github.com/golang/go/issues/25484
	return *(*string)(unsafe.Pointer(&vLen))
}

// ReadFloat64 reads 8 bytes and interprets them as a float64 IEEE 754 4 byte bit sequence.
func (f *LEBuffer) ReadFloat64() float64 {
	bits := f.ReadUint64()
	return math.Float64frombits(bits)
}

// ReadFloat32 reads 4 bytes and interprets them as a float32 IEEE 754 4 byte bit sequence.
func (f *LEBuffer) ReadFloat32() float32 {
	bits := f.ReadUint32()
	return math.Float32frombits(bits)
}

// WriteFloat32 writes a float32 IEEE 754 4 byte bit sequence.
func (f *LEBuffer) WriteFloat32(v float32) {
	bits := math.Float32bits(v)
	f.WriteUint32(bits)
}

// WriteFloat64 writes a float64 IEEE 754 8 byte bit sequence.
func (f *LEBuffer) WriteFloat64(v float64) {
	bits := math.Float64bits(v)
	f.WriteUint64(bits)
}

// Drain moves the buffer position the right amount of bytes without actually parsing it
func (f *LEBuffer) Drain(t DataType) int {
	oldPos := f.Pos
	switch t {
	case TUint8:
		f.Pos++
	case TUint16:
		f.Pos += 2
	case TUint24:
		f.Pos += 3
	case TUint32:
		f.Pos += 4
	case TUint64:
		f.Pos += 8
	case TTinyBlob:
		vLen := int(f.ReadUint8())
		f.Pos += vLen
	case TBlob:
		vLen := int(f.ReadUint16())
		f.Pos += vLen
	case TMediumBlob:
		vLen := int(f.ReadUint24())
		f.Pos += vLen
	case TLongBlob:
		vLen := int(f.ReadUint32())
		f.Pos += vLen
	case TFloat32:
		f.Pos += 4
	case TFloat64:
		f.Pos += 8
	default:
		panic("not implemented " + strconv.Itoa(int(t)))
	}
	return f.Pos - oldPos
}
