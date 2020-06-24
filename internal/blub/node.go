package blub

import (
	"bytes"
	"github.com/worldiety/ioutil"
)

type NodeReader struct {
	buf    []byte
	reader *bytes.Reader
	din    ioutil.DataInput
}

func NewNodeReader() *NodeReader {
	r := &NodeReader{
		buf:    nil,
		reader: bytes.NewReader(nil),
	}
	r.din = ioutil.NewDataInput(ioutil.LittleEndian, r.reader)
	return r
}

func (n *NodeReader) reset(buf []byte) {
	n.reader.Reset(buf)
}

// ForEachField must consume each field?
func (n *NodeReader) ForEachField(f func(name string, v interface{}) error) error {

}
