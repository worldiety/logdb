package store

import (
	"encoding/binary"
)

type Type byte

const (
	THeader     Type = 1
	TFieldTable Type = 2
	TData       Type = 3
	TTxEnd      Type = 4
)

type FieldType byte

const(
	FVarInt FieldType = 1
	FFloat32 FieldType = 2
	FInt32 FieldType = 3
	FInt64 FieldType = 4
	FString FieldType = 5
)

// Object represents a storage node (e.g. a no-sql row) in the file.
// The layout is like follows:
//   typ byte //Obj* constants
//   length varint //amount of bytes to follow until end of entry
//   fields varint //amount of fields stored by this entry
type Object []byte

// Type of the saved object
func (n Object) Type() Type {
	return n[0]
}

// Length is the amount of bytes which follows. The absolute length of an object is 1 + 4 + <value>. Negative in case
// of error.
func (n Object) Length() int64 {
	l, read := binary.Uvarint(n[1:])
	if read <= 0 {
		return int64(read)
	}
	return int64(l)
}

func (n Object) ForEachField(f func(field int64, fieldType FieldType,))error{
	l, readLength := binary.Uvarint(n[1:])
}

type Field struct {
	index uint16 // varint

}
