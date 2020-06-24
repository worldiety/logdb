package blub

import "fmt"

var EndOfNodes = fmt.Errorf("no more nodes")

// A NodeType describes the kind of the current node
type NodeType byte

const (
	NHeader     NodeType = 1
	NFieldTable NodeType = 2
	NObject     NodeType = 3
	NTxEnd      NodeType = 4
)

// A NodeReader allows to iterate sequentially over a data set.
type NodeReader2 interface {
	SetOffset(offset uint64)
	Type() NodeType
	Next() (NodeReader2, error)
}

type ObjectReader interface {
	NodeReader2
	NextField() (FieldReader, error)
}


type FieldReader interface {
}
