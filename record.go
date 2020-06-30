package logdb

import "github.com/worldiety/ioutil"

const (
	offsetRecMagic    = 0
	offsetRecSize     = 8
	offsetRecObjCount = 12
	offsetRecObjList  = 16
)

var recordMagic = [8]byte{'w', 'd', 'y', 'r', 'e', 'c', '0', '1'}

type Record struct {
	magic    [8]byte
	buf      *ioutil.LittleEndianBuffer
	size     uint32
	objCount uint32
}

func newRecord(maxSize int) *Record {
	r := &Record{buf: &ioutil.LittleEndianBuffer{
		Bytes: make([]byte, maxSize)},
		magic: recordMagic}
	r.Reset()
	return r
}

func (d *Record) MaxSize() int {
	return len(d.buf.Bytes)
}

func (d *Record) Size() uint32 {
	return d.size
}

func (d *Record) setSize(s uint32) {
	d.size = s
}

func (d *Record) ObjectCount() uint32 {
	return d.objCount
}

func (d *Record) setObjectCount(v uint32) {
	d.objCount = v
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
		_ = offsetSize // for documentation only
		size := d.buf.ReadUint24()
		d.buf.Pos -= 3
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

func (d *Record) flush() {
	d.buf.Pos = offsetRecMagic
	d.buf.WriteSlice(d.magic[:])

	d.buf.Pos = offsetRecSize
	d.buf.WriteUint32(d.size)

	d.buf.Pos = offsetRecObjCount
	d.buf.WriteUint32(d.objCount)
}

func (d *Record) reverseFlush() {
	d.buf.Pos = offsetRecSize
	d.size = d.buf.ReadUint32()

	d.buf.Pos = offsetRecObjCount
	d.objCount = d.buf.ReadUint32()
}
