package logdb

import "github.com/worldiety/ioutil"

const (
	offsetRecSize     = 0
	offsetRecObjCount = 4
	offsetRecObjList  = 8
)

type Record struct {
	buf *ioutil.LittleEndianBuffer
}

func newRecord(maxSize int) *Record {
	r := &Record{buf: &ioutil.LittleEndianBuffer{
		Bytes: make([]byte, maxSize),
	}}
	r.Reset()
	return r
}

func (d *Record) MaxSize() int {
	return len(d.buf.Bytes)
}

func (d *Record) Size() uint32 {
	d.buf.Pos = offsetRecSize
	return d.buf.ReadUint32()
}

func (d *Record) setSize(s uint32) {
	d.buf.Pos = offsetRecSize
	d.buf.WriteUint32(s)
}

func (d *Record) ObjectCount() uint32 {
	d.buf.Pos = offsetRecObjCount
	return d.buf.ReadUint32()
}

func (d *Record) setObjectCount(v uint32) {
	d.buf.Pos = offsetRecObjCount
	d.buf.WriteUint32(v)
}

func (d *Record) Bytes() []byte {
	return d.buf.Bytes[:d.Size()]
}

func (d *Record) Reset() {
	d.setSize(offsetRecObjList)
	d.setObjectCount(0)
}

func (d *Record) Add(obj *Object) {
	count := d.ObjectCount()
	size := d.Size()
	tmp := obj.Bytes()
	d.buf.Pos = int(size)
	d.buf.WriteSlice(tmp)
	d.setSize(size + uint32(len(tmp)))
	d.setObjectCount(count + 1)
}

func (d *Record) ForEach(tmp *Object, f func(offset int, object *Object) error) error {
	count := int(d.ObjectCount())
	d.buf.Pos = offsetRecObjList
	for i := 0; i < count; i++ {
		size := d.buf.ReadUint32()
		d.buf.Pos -= 4
		objBuf := d.buf.Bytes[d.buf.Pos : d.buf.Pos+int(size)]
		copy(tmp.buf.Bytes, objBuf)
		tmp.reverseFlush()
		if err := f(d.buf.Pos, tmp); err != nil {
			return err
		}
		d.buf.Pos += int(size)
	}

	return nil
}
