package logdb

import "github.com/worldiety/ioutil"

type FieldReader ioutil.TypedLittleEndianBuffer

func (f *FieldReader) ReadBlob(dst []byte) int {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadBlob(dst)
}

func (f *FieldReader) ReadMutableString(dst []byte) string {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadString(dst)
}

func (f *FieldReader) ReadInt() int64 {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadInt()
}

func (f *FieldReader) ReadFloat() float64 {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadFloat()
}

func (f *FieldReader) ReadUint8() uint8 {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadUint8()
}

func (f *FieldReader) ReadInt8() int8 {
	return int8((*ioutil.TypedLittleEndianBuffer)(f).ReadUint8())
}

func (f *FieldReader) ReadUint16() uint16 {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadUint16()
}

func (f *FieldReader) ReadInt16() int16 {
	return int16((*ioutil.TypedLittleEndianBuffer)(f).ReadUint16())
}

func (f *FieldReader) ReadUint24() uint32 {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadUint24()
}

func (f *FieldReader) ReadInt24() int32 {
	return int32((*ioutil.TypedLittleEndianBuffer)(f).ReadUint24())
}

func (f *FieldReader) ReadUint32() uint32 {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadUint32()
}

func (f *FieldReader) ReadInt32() int32 {
	return int32((*ioutil.TypedLittleEndianBuffer)(f).ReadUint32())
}

func (f *FieldReader) ReadUint64() uint64 {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadUint64()
}

func (f *FieldReader) ReadInt64() int64 {
	return int64((*ioutil.TypedLittleEndianBuffer)(f).ReadUint64())
}

func (f *FieldReader) ReadBlob8(dst []byte) int {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadBlob8(dst)
}

func (f *FieldReader) ReadBlob16(dst []byte) int {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadBlob16(dst)
}

func (f *FieldReader) ReadBlob24(dst []byte) int {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadBlob24(dst)
}

func (f *FieldReader) ReadBlob32(dst []byte) int {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadBlob32(dst)
}

func (f *FieldReader) ReadFloat32() float32 {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadFloat32()
}

func (f *FieldReader) ReadFloat64() float64 {
	return (*ioutil.TypedLittleEndianBuffer)(f).ReadFloat64()
}
