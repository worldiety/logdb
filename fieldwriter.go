package logdb

import "github.com/worldiety/ioutil"

type FieldWriter ioutil.TypedLittleEndianBuffer

func (f *FieldWriter) WriteBlob(dst []byte) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteBlob(dst)
}

func (f *FieldWriter) WriteString(str string) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteString(str)
}

func (f *FieldWriter) WriteInt(v int64) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteInt(v)
}

func (f *FieldWriter) WriteFloat(v float64) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteFloat(v)
}

func (f *FieldWriter) WriteUint8(v uint8) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteUint8(v)
}

func (f *FieldWriter) WriteUint16(v uint16) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteUint16(v)
}

func (f *FieldWriter) WriteUint24(v uint32) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteUint24(v)
}

func (f *FieldWriter) WriteUint32(v uint32) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteUint32(v)
}

func (f *FieldWriter) WriteUint64(v uint64) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteUint64(v)
}

func (f *FieldWriter) WriteFloat32(v float32) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteFloat32(v)
}

func (f *FieldWriter) WriteFloat64(v float64) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteFloat64(v)
}

func (f *FieldWriter) WriteBlob8(v []byte) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteBlob8(v)
}

func (f *FieldWriter) WriteBlob16(v []byte) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteBlob16(v)
}

func (f *FieldWriter) WriteBlob24(v []byte) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteBlob24(v)
}

func (f *FieldWriter) WriteBlob32(v []byte) {
	(*ioutil.TypedLittleEndianBuffer)(f).WriteBlob32(v)
}
