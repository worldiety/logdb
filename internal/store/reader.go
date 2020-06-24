package store

import "github.com/worldiety/ioutil"

type NodeReader struct {
	din ioutil.DataInput
}

func (n NodeReader) Next() bool{

}

func (n NodeReader) Type() Type {
	return n.din.ReadInt8()
}

type ObjectReader struct{
	din ioutil.DataInput
}

type FieldReader struct{
	din ioutil.DataInput
}
