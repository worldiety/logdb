package blub

import (
	"bytes"
	"github.com/worldiety/ioutil"
	"io"
)

type Page struct {
	buf        []byte
	reader     *bytes.Reader
	din        ioutil.DataInput
	nodeReader *NodeReader
}

func NewPage(size int) *Page {
	p := &Page{
		buf: make([]byte, size),
	}
	p.reader = bytes.NewReader(p.buf)
	p.din = ioutil.NewDataInput(ioutil.LittleEndian, p.reader)
	p.nodeReader = NewNodeReader()
	return p
}

func (p *Page) clear() {
	for i := range p.buf {
		p.buf[i] = 0
	}
}

func (p *Page) Size() int {
	return len(p.buf)
}

func (p *Page) ForEachNode(f func(reader *NodeReader) error) error {
	p.reader.Reset(p.buf)

	nc := p.din.ReadUvarint()
	for i := uint64(0); i < nc; i++ {
		nodeSize := int64(p.din.ReadUvarint())
		pos, _ := p.reader.Seek(0, io.SeekCurrent)
		nodeBuf := p.buf[pos : pos+nodeSize]
		p.nodeReader.reset(nodeBuf)
		err := f(p.nodeReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Page) NodeCount() (uint64, error) {
	p.reader.Reset(p.buf)
	return p.din.ReadUvarint(), p.din.Error()
}
